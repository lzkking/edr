package plugin

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/k0kubun/pp"
	"github.com/lzkking/edr/agent/internal/agent"
	"github.com/lzkking/edr/agent/internal/buffer"
	"github.com/lzkking/edr/agent/utils"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func (p *Plugin) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.shutdown = true
	if p.IsExited() {
		return
	}

	p.Infof("plugin is running ,will shutdown it")
	p.tx.Close()
	p.rx.Close()
	select {
	case <-time.After(time.Second * 10):
		p.Warn("because of plugin exit's timout,will kill it")
		syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
		<-p.done
		p.Info("plugin has been killed")
	case <-p.done:
		p.Info("plugin has been shutdown gracefully")
	}
}

func Load(ctx context.Context, config pb.ConfigItem) (plg *Plugin, err error) {
	if validName.FindString(regexp.QuoteMeta(config.Name)) == "" {
		return nil, fmt.Errorf("invalid config name: %s", config.Name)
	}

	loadedPlg, ok := m.Load(config.Name)
	if ok {
		loadedPlg := loadedPlg.(*Plugin)
		if loadedPlg.Config.Version == config.Version && loadedPlg.cmd.ProcessState == nil {
			err = ErrDuplicatePlugin
			return
		}

		if loadedPlg.Config.Version != config.Version && loadedPlg.cmd.ProcessState == nil {
			loadedPlg.Infof("because of the different plugin's version,the previous version will been shutdown...")
			loadedPlg.Shutdown()
			loadedPlg.Infof("shutdown successfully")
		}
	}

	if config.Signature == "" {
		config.Signature = config.SHA256
	}

	logger := zap.S().With("plugin", config.Name, "pver", config.Version, "psign", config.Signature)
	logger.Infof("plugin is loading...")

	workingDirectory := path.Join(agent.WorkDirectory, "plugin", config.Name)
	patternDirectory := path.Join(agent.WorkDirectory, "plugin", "*")
	match, err := path.Match(patternDirectory, workingDirectory)
	if match != true {
		logger.Warn("invalid path & name for plugin: ", config.Name)
		return
	}

	os.Remove(path.Join(workingDirectory, config.Name+".stderr"))
	os.Remove(path.Join(workingDirectory, config.Name+".stdout"))
	execPath := path.Join(workingDirectory, config.Name)
	err = utils.CheckSignature(execPath, config.Signature)
	if err != nil {
		logger.Warn("check local plugin's signature failed: ", err)
		logger.Infof("downloading plugin from remote server...")
		err = utils.Download(ctx, execPath, config)
		if err != nil {
			return
		}
		logger.Infof("download done")
	}

	cmd := exec.Command(execPath)
	var rx_r, rx_w, tx_r, tx_w *os.File
	rx_r, rx_w, err = os.Pipe()
	if err != nil {
		return
	}
	tx_r, tx_w, err = os.Pipe()
	if err != nil {
		return
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.ExtraFiles = append(cmd.ExtraFiles, tx_r, rx_w)
	cmd.Dir = workingDirectory
	var errFile *os.File
	errFile, err = os.OpenFile(execPath+".stderr", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o0600)
	if err != nil {
		return
	}
	defer errFile.Close()
	cmd.Stderr = errFile
	if config.Detail != "" {
		cmd.Env = append(cmd.Env, "DETAIL="+config.Detail)
	}
	logger.Info("plugin's process will start")
	err = cmd.Start()
	tx_r.Close()
	rx_w.Close()
	if err != nil {
		return
	}
	plg = &Plugin{
		Config:        config,
		mu:            &sync.Mutex{},
		cmd:           cmd,
		rx:            rx_r,
		updateTime:    time.Now(),
		reader:        bufio.NewReaderSize(rx_r, 1024*128),
		tx:            tx_w,
		done:          make(chan struct{}),
		taskCh:        make(chan pb.PluginTask),
		wg:            &sync.WaitGroup{},
		SugaredLogger: logger,
	}
	plg.wg.Add(3)

	//处理插件进程的关闭
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of waiting plugin's process will exit")
		err = cmd.Wait()
		rx_r.Close()
		tx_w.Close()

		if err != nil {
			plg.Error("plugin has exited with error: %v, code: %d", err, cmd.ProcessState.ExitCode())
		} else {
			plg.Infof("plugin has exited with code %d", cmd.ProcessState.ExitCode())
		}

		close(plg.done)
	}()

	//	接收插件传递来的数据
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of waiting plugin's process will exit")
		for {
			zap.S().Debugf("接收插件数据")
			rec, err := plg.ReceiveData()
			if err != nil {
				zap.S().Debugf("接收插件数据失败")
				if errors.Is(err, bufio.ErrBufferFull) {
					plg.Warn("when receiving data, buffer is full, skip this record")
					continue
				} else if !(errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrClosed)) {
					plg.Error("when receiving data, an error occurred: ", err)
				} else {
					break
				}
			}
			zap.S().Debugf("%v", pp.Sprintf("%v", rec))

			buffer.WriteEncodedRecord(rec)
		}
	}()

	//向插件发送数据
	go func() {
		defer plg.wg.Done()
		defer plg.Info("gorountine of waiting plugin's process will exit")
		for {
			select {
			case <-plg.done:
				return
			case task := <-plg.taskCh:
				data, err := proto.Marshal(&task)
				if err != nil {
					plg.Errorf("when marshaling a task, an error occurred: %v, ignored this task: %+v", err, task)
					continue
				}
				var dst = make([]byte, 4+len(data))
				copy(dst[4:], data)

				s := proto.Size(&task)

				binary.LittleEndian.PutUint32(dst[:4], uint32(s))
				var n int
				n, err = plg.tx.Write(dst)
				if err != nil {
					if !errors.Is(err, os.ErrClosed) {
						plg.Error("when sending task, an error occurred: ", err)
					}
					return
				}
				atomic.AddUint64(&plg.rxCnt, 1)
				atomic.AddUint64(&plg.rxBytes, uint64(n))
			}
		}
	}()
	m.Store(config.Name, plg)
	return
}
