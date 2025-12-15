package sdkgenerator

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	generators map[SDK]Generator
}

type RegistryBuilder struct {
	runner     CommandRunner
	generators []Generator
}

func NewRegistryBuilder(runner CommandRunner) *RegistryBuilder {
	return &RegistryBuilder{
		runner:     runner,
		generators: []Generator{},
	}
}

func (b *RegistryBuilder) WithGenerator(g Generator) *RegistryBuilder {
	b.generators = append(b.generators, g)
	return b
}

func (b *RegistryBuilder) WithDefaultGenerators() *RegistryBuilder {
	b.generators = append(b.generators,
		NewGoProtobufGenerator(b.runner),
		NewGoConnectRpcGenerator(b.runner),
		NewGoGrpcGenerator(b.runner),
		NewJsBufbuildEsGenerator(b.runner),
		NewJsProtobufGenerator(b.runner),
		NewJsConnectRpcGenerator(b.runner),
	)
	return b
}

func (b *RegistryBuilder) Build() *Registry {
	r := &Registry{
		generators: make(map[SDK]Generator),
	}
	for _, gen := range b.generators {
		r.Register(gen)
	}
	return r
}

func NewRegistry(runner CommandRunner) *Registry {
	return NewRegistryBuilder(runner).WithDefaultGenerators().Build()
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
