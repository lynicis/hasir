package sdkgenerator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry(nil)
	assert.NotNil(t, r)

	expectedSdks := []SDK{
		SdkGoProtobuf,
		SdkGoConnectRpc,
		SdkGoGrpc,
		SdkJsBufbuildEs,
		SdkJsProtobuf,
		SdkJsConnectrpc,
		SdkRustProtobuf,
		SdkRustGrpc,
		SdkJavaProtobuf,
		SdkJavaGrpc,
		SdkCsharpProtobuf,
		SdkCsharpGrpc,
	}

	for _, sdk := range expectedSdks {
		g, err := r.Get(sdk)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, sdk, g.SDK())
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry(NewMockCommandRunner())

	t.Run("existing generator", func(t *testing.T) {
		g, err := r.Get(SdkGoProtobuf)
		require.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, SdkGoProtobuf, g.SDK())
	})

	t.Run("non-existing generator", func(t *testing.T) {
		g, err := r.Get(SDK("UNKNOWN"))
		require.Error(t, err)
		assert.Nil(t, g)
		assert.Contains(t, err.Error(), "no generator registered")
	})
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry(NewMockCommandRunner())

	sdks := r.List()
	assert.Len(t, sdks, 12)

	sdkMap := make(map[SDK]bool)
	for _, sdk := range sdks {
		sdkMap[sdk] = true
	}

	assert.True(t, sdkMap[SdkGoProtobuf])
	assert.True(t, sdkMap[SdkGoConnectRpc])
	assert.True(t, sdkMap[SdkGoGrpc])
	assert.True(t, sdkMap[SdkJsBufbuildEs])
	assert.True(t, sdkMap[SdkJsProtobuf])
	assert.True(t, sdkMap[SdkJsConnectrpc])
	assert.True(t, sdkMap[SdkRustProtobuf])
	assert.True(t, sdkMap[SdkRustGrpc])
	assert.True(t, sdkMap[SdkJavaProtobuf])
	assert.True(t, sdkMap[SdkJavaGrpc])
	assert.True(t, sdkMap[SdkCsharpProtobuf])
	assert.True(t, sdkMap[SdkCsharpGrpc])
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry(NewMockCommandRunner())

	mockGenerator := &mockGenerator{sdk: SDK("CUSTOM")}
	r.Register(mockGenerator)

	g, err := r.Get(SDK("CUSTOM"))
	require.NoError(t, err)
	assert.Equal(t, SDK("CUSTOM"), g.SDK())
}

type mockGenerator struct {
	sdk     SDK
	dirName string
}

func (m *mockGenerator) Generate(_ context.Context, _ GeneratorInput) (*GeneratorOutput, error) {
	return &GeneratorOutput{}, nil
}

func (m *mockGenerator) SDK() SDK {
	return m.sdk
}

func (m *mockGenerator) DirName() string {
	if m.dirName != "" {
		return m.dirName
	}
	return "custom"
}

func (m *mockGenerator) Validate(_ GeneratorInput) error {
	return nil
}

func TestRegistryBuilder(t *testing.T) {
	runner := NewMockCommandRunner()

	t.Run("build registry with custom generators", func(t *testing.T) {
		customGen := &mockGenerator{sdk: SDK("CUSTOM"), dirName: "custom-sdk"}

		r := NewRegistryBuilder(runner).
			WithGenerator(NewGoProtobufGenerator(runner)).
			WithGenerator(customGen).
			Build()

		g1, err := r.Get(SdkGoProtobuf)
		require.NoError(t, err)
		assert.NotNil(t, g1)

		g2, err := r.Get(SDK("CUSTOM"))
		require.NoError(t, err)
		assert.NotNil(t, g2)
		assert.Equal(t, "custom-sdk", g2.DirName())
	})

	t.Run("build registry with default generators", func(t *testing.T) {
		r := NewRegistryBuilder(runner).
			WithDefaultGenerators().
			Build()

		sdks := r.List()
		assert.Len(t, sdks, 12)
	})

	t.Run("build registry with mix of default and custom", func(t *testing.T) {
		customGen := &mockGenerator{sdk: SDK("CUSTOM"), dirName: "custom"}

		r := NewRegistryBuilder(runner).
			WithDefaultGenerators().
			WithGenerator(customGen).
			Build()

		sdks := r.List()
		assert.Len(t, sdks, 13)

		g, err := r.Get(SDK("CUSTOM"))
		require.NoError(t, err)
		assert.Equal(t, "custom", g.DirName())
	})
}

func TestNewBufRegistry(t *testing.T) {
	runner := NewMockCommandRunner()
	registry := NewBufRegistry(runner)

	sdks := registry.List()
	assert.Len(t, sdks, 12)

	for _, sdk := range []SDK{SdkGoProtobuf, SdkGoConnectRpc, SdkGoGrpc, SdkJsBufbuildEs, SdkJsProtobuf, SdkJsConnectrpc, SdkRustProtobuf, SdkRustGrpc, SdkJavaProtobuf, SdkJavaGrpc, SdkCsharpProtobuf, SdkCsharpGrpc} {
		g, err := registry.Get(sdk)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, sdk, g.SDK())
	}
}

func TestRegistryBuilder_WithBufGenerators(t *testing.T) {
	runner := NewMockCommandRunner()
	registry := NewRegistryBuilder(runner).WithBufGenerators().Build()

	sdks := registry.List()
	assert.Len(t, sdks, 12)
}
