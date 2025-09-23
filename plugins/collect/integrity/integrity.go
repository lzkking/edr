package integrity

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"github.com/juju/ratelimit"
	"github.com/lzkking/edr/plugins/collect/engine"
	"github.com/lzkking/edr/plugins/collect/rpm"
	plugins "github.com/lzkking/edr/plugins/lib"
	hash2 "hash"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxVerifyFileSize = 100 * 1024 * 1024

func isBin(path string) bool {
	d := filepath.Dir(path)
	return filepath.Ext(path) == "" &&
		(d == "/bin" || d == "/sbin" ||
			d == "/usr/bin" || d == "/usr/sbin" ||
			d == "/usr/local/bin" || d == "/usr/local/sbin")
}

type IntegrityHandler struct{}

func (h *IntegrityHandler) Name() string {
	return "integrity"
}

func (h *IntegrityHandler) DataType() int {
	return 7317
}

func (h *IntegrityHandler) Handle(c *plugins.Client, cache *engine.Cache, seq string) {
	if db, err := os.Open("/var/lib/dpkg/status"); err == nil {
		nvm := map[string]string{}
		s := bufio.NewScanner(io.LimitReader(db, 25*1024*1024))
		s.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			if i := strings.Index(string(data), "\nPackage: "); i >= 0 {
				return i + 1, data[0:i], nil
			}
			if atEOF {
				return len(data), data, nil
			}
			return
		})
		for s.Scan() {
			lines := strings.Split(s.Text(), "\n")
			var n, v string
			for _, line := range lines {
				fields := strings.SplitN(line, ": ", 2)
				if len(fields) != 2 {
					continue
				}
				switch fields[0] {
				case "Package":
					n = fields[1]
				case "Version":
					v = fields[1]
				}
				if n != "" && v != "" {
					break
				}
			}
			if n != "" && v != "" {
				nvm[n] = v
			}
		}
		db.Close()

		if entries, err := os.ReadDir("/var/lib/dpkg/info"); err == nil {
			//paths
			ps := []string{}
			for _, entry := range entries {
				if filepath.Ext(entry.Name()) == ".md5sums" {
					ps = append(ps, filepath.Join("/var/lib/dpkg/info", entry.Name()))
				}
			}
			hash := md5.New()
			for _, p := range ps {
				pn := strings.TrimSuffix(filepath.Base(p), ".md5sums")
				if v, ok := nvm[pn]; ok {
					fs := map[string]string{}
					if f, err := os.Open(p); err == nil {
						s := bufio.NewScanner(f)
						for s.Scan() {
							fds := strings.Fields(s.Text())
							if len(fds) == 2 {
								p := filepath.Join("/", fds[1])
								if isBin(p) {
									fs[p] = fds[0]
								}
							}
						}
						f.Close()
						for path, odigest := range fs {
							if f, err := os.Open(path); err == nil {
								if st, err := f.Stat(); err == nil && st.Size() < maxVerifyFileSize {
									r := ratelimit.Reader(f, ratelimit.NewBucketWithRate(1024*1024, 1024*1024))
									if _, err = io.Copy(hash, r); err == nil {
										digest := hex.EncodeToString(hash.Sum(nil))
										hash.Reset()
										if digest != odigest {
											c.SendRecord(&plugins.Record{
												DataType:  int32(h.DataType()),
												Timestamp: time.Now().Unix(),
												Data: &plugins.Payload{
													Fields: map[string]string{
														"software_name":    pn,
														"digest":           digest,
														"origin_digest":    odigest,
														"digest_algorithm": "md5",
														"exe":              path,
														"modify_time":      strconv.FormatInt(st.ModTime().Unix(), 10),
														"software_version": v,
														"package_seq":      seq,
													},
												},
											})
											cache.Put(h.DataType(), path, nil)
										}
									}
								}
								f.Close()
							}
						}
					}
				}
			}
		}
	} else if db, err := rpm.OpenDatabase(); err == nil {
		db.WalkPackages(func(p rpm.Package) {
			var hash hash2.Hash
			switch p.DigestAlgorithm {
			case rpm.PGPHASHALGO_MD5:
				hash = md5.New()
			case rpm.PGPHASHALGO_SHA1:
				hash = sha1.New()
			case rpm.PGPHASHALGO_SHA256:
				hash = sha256.New()
			case rpm.PGPHASHALGO_SHA512:
				hash = sha512.New()
			default:
				return
			}
			for _, f := range p.Files {
				if isBin(f.Path) && f.Digest != "" {
					if af, err := os.Open(f.Path); err == nil {
						if st, err := af.Stat(); err == nil && st.Size() < maxVerifyFileSize {
							r := ratelimit.Reader(af, ratelimit.NewBucketWithRate(1024*1024, 1024*1024))
							hash.Reset()
							if _, err = io.Copy(hash, r); err == nil {
								digest := hex.EncodeToString(hash.Sum(nil))
								if digest != f.Digest {
									c.SendRecord(&plugins.Record{
										DataType:  int32(h.DataType()),
										Timestamp: time.Now().Unix(),
										Data: &plugins.Payload{
											Fields: map[string]string{
												"software_name":    p.Name,
												"digest":           digest,
												"origin_digest":    f.Digest,
												"digest_algorithm": p.DigestAlgorithm.String(),
												"exe":              f.Path,
												"modify_time":      strconv.FormatInt(st.ModTime().Unix(), 10),
												"software_version": p.Version,
												"package_seq":      seq,
											},
										},
									})
									cache.Put(h.DataType(), f.Path, nil)
								}
							}
						}
						af.Close()
					}
				}
			}
		})
		db.Close()
	}
}
