package sdkgenerator

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	generators map[SDK]Generator
}

func NewRegistry(runner CommandRunner) *Registry {
	r := &Registry{
		generators: make(map[SDK]Generator),
	}

	r.Register(NewGoProtobufGenerator(runner))
	r.Register(NewGoConnectRpcGenerator(runner))
	r.Register(NewGoGrpcGenerator(runner))
	r.Register(NewJsBufbuildEsGenerator(runner))
	r.Register(NewJsProtobufGenerator(runner))
	r.Register(NewJsConnectRpcGenerator(runner))

	return r
}

func (r *Registry) Register(g Generator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[g.SDK()] = g
}

func (r *Registry) Get(sdk SDK) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.generators[sdk]
	if !ok {
		return nil, fmt.Errorf("no generator registered for SDK: %s", sdk)
	}

	return g, nil
}

func (r *Registry) List() []SDK {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sdks := make([]SDK, 0, len(r.generators))
	for sdk := range r.generators {
		sdks = append(sdks, sdk)
	}

	return sdks
}
