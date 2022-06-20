package kv

import "bytes"

func New() *KV {
	return &KV{mp: make(map[string][]byte)}
}

func (a *KV) Close() error {
	return nil
}

func (a *KV) Del(k string) error {
	a.Lock()
	defer a.Unlock()
	delete(a.mp, k)
	return nil
}

func (a *KV) Set(k string, v []byte) error {
	a.Lock()
	defer a.Unlock()
	a.mp[k] = v
	return nil
}

func (a *KV) Get(k string, buf *bytes.Buffer) ([]byte, error) {
	a.Lock()
	defer a.Unlock()
	v, ok := a.mp[k]
	if !ok {
		return nil, ErrNotExist
	}
	buf.Reset()
	if len(v) > buf.Cap() {
		buf.Grow(len(v))
	}
	data := buf.Bytes()[:len(v)]
	copy(data, v)
	return data, nil
}

func (a *KV) Range() ([]string, [][]byte) {
	a.Lock()
	defer a.Unlock()
	names := make([]string, 0, len(a.mp))
	datas := make([][]byte, 0, len(a.mp))
	for k, v := range a.mp {
		names = append(names, k)
		datas = append(datas, v)
	}
	return names, datas
}
