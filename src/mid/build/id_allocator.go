package build

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type BeanIdAllocator interface {
	Alloc(key string) int
	Output() error
}

type IdPair struct {
	Key string
	Id  int
}

func JoinBeanKey(pkgName, beanName string) string {
	return pkgName + "." + beanName
}

func SplitBeanKey(key string) (pkgName, beanName string) {
	index := strings.LastIndex(key, ".")
	if index >= 0 {
		return key[:index], key[index+1:]
	}
	return "", key
}

func NewBeanIdAllocator(name, opts string) (BeanIdAllocator, error) {
	switch name {
	case "file":
		return NewFileBeanIdAllocator(opts)
	default:
		return nil, errors.New("unsupported bean id allocator: " + name)
	}
}

// fileBeanIdAllocator implements BeanIdAllocator
type fileBeanIdAllocator struct {
	filename string
	perm     os.FileMode
	idMap    map[string]int
	ids      []IdPair
	maxId    int
}

func NewFileBeanIdAllocator(filename string) (BeanIdAllocator, error) {
	allocator := &fileBeanIdAllocator{
		filename: filename,
		idMap:    make(map[string]int),
	}
	if err := allocator.readInit(); err != nil {
		return nil, err
	}
	return allocator, nil
}

func (allocator *fileBeanIdAllocator) Alloc(key string) int {
	if id, found := allocator.idMap[key]; found {
		return id
	}
	allocator.maxId += 1 + rand.Intn(5)
	id := allocator.maxId
	allocator.idMap[key] = id
	allocator.ids = append(allocator.ids, IdPair{Key: key, Id: id})
	return id
}

func (allocator *fileBeanIdAllocator) Output() error {
	w, err := os.OpenFile(allocator.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, allocator.perm)
	if err != nil {
		return err
	}
	defer w.Close()
	if len(allocator.ids) == 0 {
		return nil
	}
	sort.Slice(allocator.ids, func(i, j int) bool { return allocator.ids[i].Id < allocator.ids[j].Id })
	for _, pair := range allocator.ids {
		fmt.Fprintf(w, "%s=%d\n", pair.Key, pair.Id)
	}
	return nil
}

func (allocator *fileBeanIdAllocator) readInit() error {
	info, err := os.Stat(allocator.filename)
	if err != nil {
		if os.IsNotExist(err) {
			allocator.perm = 0666
			return nil
		}
		return err
	}
	if info.IsDir() {
		return errors.New(allocator.filename + " is not a regular file")
	}
	allocator.perm = info.Mode()
	file, err := os.Open(allocator.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	idMap, err := ReadBeanIds(file)
	if err != nil {
		return err
	}
	allocator.idMap = idMap
	for key, id := range idMap {
		if len(allocator.ids) == 0 || allocator.maxId < id {
			allocator.maxId = id
		}
		allocator.ids = append(allocator.ids, IdPair{Key: key, Id: id})
	}
	return nil
}

func ReadBeanIds(reader io.Reader) (map[string]int, error) {
	var (
		idMap   = make(map[string]int)
		advance int
		token   []byte
	)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	count := 0
	for advance < len(data) {
		count++
		lineno := fmt.Sprintf("line %d: ", count)
		total := advance
		advance, token, _ = bufio.ScanLines(data[advance:], true)
		if advance == 0 {
			break
		}
		advance = total + advance
		tok := strings.TrimSpace(string(token))
		if tok == "" {
			continue
		}
		kv := strings.SplitN(tok, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New(lineno + string(token) + " is not a key value pair seperated by =")
		}
		key, value := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		id, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New(lineno + value + " is not a integer")
		}
		idMap[key] = id
	}
	return idMap, nil
}
