package smarthome

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"sort"
	"sync"
)

type Light struct {
	Name string
	On   bool
}

type LightClient struct {
	data    map[string]*Light
	dataMux sync.Mutex
}

func newLightClient() *LightClient {
	lc := &LightClient{
		data: map[string]*Light{},
	}
	js, _ := ioutil.ReadFile("/tmp/godays2020/lights.json")
	var lights []Light
	_ = json.Unmarshal(js, &lights)

	for _, light := range lights {
		l := lc.getLight(light.Name)
		l.On = light.On
	}
	return lc
}

func (lc *LightClient) close() {
	lights, _ := lc.List(nil)
	js, _ := json.Marshal(lights)
	_ = ioutil.WriteFile("/tmp/godays2020/lights.json", js, 0700)
}

func (lc *LightClient) getLight(name string) *Light {
	if l, ok := lc.data[name]; ok {
		return l
	}
	lc.data[name] = &Light{Name: name}
	return lc.data[name]
}

func (lc *LightClient) Switch(ctx context.Context, name string, on bool) error {
	lc.dataMux.Lock()
	defer lc.dataMux.Unlock()

	light := lc.getLight(name)
	light.On = on
	return nil
}

func (lc *LightClient) Get(ctx context.Context, name string) (Light, error) {
	lc.dataMux.Lock()
	defer lc.dataMux.Unlock()

	light := lc.getLight(name)
	return *light, nil
}

func (lc *LightClient) List(ctx context.Context) ([]Light, error) {
	lc.dataMux.Lock()
	defer lc.dataMux.Unlock()

	var lights lightsByName
	for _, light := range lc.data {
		lights = append(lights, *light)
	}

	sort.Sort(lights)
	return lights, nil
}

// lightsByName sorts Lights by name
type lightsByName []Light

func (items lightsByName) Len() int {
	return len(items)
}

func (items lightsByName) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items lightsByName) Less(i, j int) bool {
	return items[i].Name < items[j].Name
}
