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
	assert.Len(t, sdks, 6)

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
	sdk SDK
}

func (m *mockGenerator) Generate(_ context.Context, _ GeneratorInput) (*GeneratorOutput, error) {
	return &GeneratorOutput{}, nil
}

func (m *mockGenerator) SDK() SDK {
	return m.sdk
}

func (m *mockGenerator) Validate(_ GeneratorInput) error {
	return nil
}
