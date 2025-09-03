package agent

import (
	"bytes"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/lzkking/edr/agent/config"
	"os"
	"path/filepath"
)

var (
	Context, Cancel = context.WithCancel(context.Background())
	ID              = ""
	WorkDirectory   = ""
	Product         = "rock-agent"
	Version         string
)

func fromUUIDFile(file string) (id uuid.UUID, err error) {
	var idBytes []byte
	idBytes, err = os.ReadFile(file)
	if err == nil {
		id, err = uuid.ParseBytes(bytes.TrimSpace(idBytes))
	}
	return
}
func fromIDFile(file string) (id []byte, err error) {
	id, err = os.ReadFile(file)
	if err == nil {
		if len(id) < 6 {
			err = errors.New("id too short")
			return
		}
		id = bytes.TrimSpace(id)
	}
	return
}

func init() {
	agentConfig := config.GetServerConfig()
	workDir := agentConfig.WorkDir
	if workDir == "" {
		WorkDirectory = "/var/run"
	} else {
		WorkDirectory = workDir
	}

	if _, err := os.Stat(WorkDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(WorkDirectory, os.ModePerm)
		if err != nil {
			panic("创建工作文件夹失败")
		}
	}

	defer func() {
		os.WriteFile(filepath.Join(WorkDirectory, "machine-id"), []byte(ID), 0600)
	}()

	mid, err := fromUUIDFile(filepath.Join(WorkDirectory, "machine-id"))
	if err == nil {
		ID = mid.String()
		return
	}

	source := []byte{}
	isid, err := fromIDFile("/var/lib/cloud/data/instance-id")
	if err == nil {
		source = append(source, isid...)
	}
	pdid, err := fromIDFile("/sys/class/dmi/id/product_uuid")
	if err == nil {
		source = append(source, pdid...)
	}
	emac, err := fromIDFile("/sys/class/net/eth0/address")
	if err == nil {
		source = append(source, emac...)
	}
	if len(source) > 8 &&
		string(pdid) != "03000200-0400-0500-0006-000700080009" &&
		string(pdid) != "02000100-0300-0400-0005-000600070008" {
		pname, err := fromIDFile("/sys/class/dmi/id/product_name")
		if err == nil && len(pname) != 0 &&
			!bytes.Equal(pname, []byte("--")) &&
			!bytes.Equal(pname, []byte("unknown")) &&
			!bytes.Equal(pname, []byte("To be filled by O.E.M.")) &&
			!bytes.Equal(pname, []byte("OEM not specify")) &&
			!bytes.Equal(bytes.ToLower(pname), []byte("t.b.d")) {
			ID = uuid.NewSHA1(uuid.NameSpaceOID, source).String()
		}
		return
	}
	mid, err = fromUUIDFile("/etc/machine-id")
	if err == nil {
		ID = mid.String()
		return
	}
	if err.Error() == "invalid UUID format" {
		source, err := fromIDFile("/etc/machine-id")
		if err == nil {
			ID = uuid.NewSHA1(uuid.NameSpaceOID, source).String()
			return
		}
	}

	ID = uuid.New().String()
}
