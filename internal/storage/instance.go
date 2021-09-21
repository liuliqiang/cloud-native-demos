package storage

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "sync"

    log "github.com/liuliqiang/log4go"
)

type Instance struct {
    Name   string `json:"name"`
    Health bool   `json:"health"`
    Host   string `json:"host"`
    Count  int    `json:"count"`
}

type InstanceStorage interface {
    AddInstance(inst *Instance) (err error)
    GetInstance(name string) (*Instance, error)
    ListInstance() ([]*Instance, error)
    RemoveInstance(name string) (err error)
}

var (
    _ InstanceStorage = (*fileInstanceStorage)(nil)
)

type fileInstanceStorage struct {
    file  string
    cache sync.Map
}

func NewFileInstanceStorage(dir string) (InstanceStorage, error) {
    file := path.Join(dir, "inst.data")
    fis := &fileInstanceStorage{
        file:  file,
        cache: sync.Map{},
    }

    if err := fis.readFile(); err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }

    return fis, nil
}

func (f *fileInstanceStorage) AddInstance(inst *Instance) (err error) {
    f.cache.Store(inst.Name, inst)

    if err = f.updateFile(); err != nil {
        return fmt.Errorf("persistence cache: %v", err)
    }
    return nil
}

var (
    ErrNotFound = errors.New("not found")
)

func (f *fileInstanceStorage) GetInstance(name string) (*Instance, error) {
    inst, exists := f.cache.Load(name)
    if !exists {
        return nil, ErrNotFound
    }

    return inst.(*Instance), nil
}

func (f *fileInstanceStorage) ListInstance() ([]*Instance, error) {
    var inst []*Instance
    f.cache.Range(func(key, v interface{}) bool {
        inst = append(inst, v.(*Instance))
        return true
    })

    return inst, nil
}

func (f *fileInstanceStorage) RemoveInstance(name string) (err error) {
    if _, exists := f.cache.Load(name); exists {
        f.cache.Delete(name)
    }

    if err = f.updateFile(); err != nil {
        return fmt.Errorf("persistence: %v", err)
    }
    return nil
}

const (
    DefaultFileMode = 0644
)

func (f *fileInstanceStorage) updateFile() (err error) {
    var insts []*Instance
    f.cache.Range(func(key, v interface{}) bool {
        insts = append(insts, v.(*Instance))
        return true
    })

    bytes, err := json.Marshal(insts)
    if err != nil {
        return fmt.Errorf("marshal cache: %v", err)
    }

    if err = ioutil.WriteFile(f.file, bytes, DefaultFileMode); err != nil {
        return fmt.Errorf("write file: %v", err)
    }

    return nil
}

func (f *fileInstanceStorage) readFile() (err error) {
    fd, err := os.OpenFile(f.file, os.O_CREATE|os.O_RDWR, DefaultFileMode)
    if err != nil {
        return fmt.Errorf("open file: %v", err)
    }

    bytes, err := ioutil.ReadAll(fd)
    if err != nil {
        return fmt.Errorf("read file: %v", err)
    }

    if len(bytes) == 0 {
        log.Info("Create storage file.")
        return nil
    }

    var insts []*Instance
    if err = json.Unmarshal(bytes, &insts); err != nil {
        return fmt.Errorf("unmarshal file: %v", err)
    }

    for _, inst := range insts {
        f.cache.Store(inst.Name, inst)
    }

    return nil
}
