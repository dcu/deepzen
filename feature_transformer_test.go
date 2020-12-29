package tabnet

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func TestFeatureTransformer(t *testing.T) {
	testCases := []struct {
		desc              string
		input             tensor.Tensor
		vbs               int
		independentBlocks int
		output            int
		expectedShape     tensor.Shape
		expectedErr       string
		expectedOutput    []float64
	}{
		{
			desc: "",
			input: tensor.New(
				tensor.WithShape(6, 2),
				tensor.WithBacking([]float64{0.4, 1.4, 2.4, 3.4, 4.4, 5.4, 6.4, 7.4, 8.4, 9.4, 10.4, 11.4}),
			),
			vbs:               2,
			output:            2,
			independentBlocks: 5,
			expectedShape:     tensor.Shape{6, 2},
			expectedOutput:    []float64{-0.46379327241699114, -0.2870165771203542, 1.87719497530188, 2.053971670598517, 0.24331350876955662, 0.42009020406619374, 2.5843017564884283, 2.761078451785065, 0.9504202899561044, 1.1271969852527415, 3.291408537674975, 3.468185232971612},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.desc, func(t *testing.T) {
			c := require.New(t)

			g := NewGraph()
			tn := &Model{g: g}

			input := NewTensor(g, tensor.Float64, tcase.input.Dims(), WithShape(tcase.input.Shape()...), WithName("input"), WithValue(tcase.input))
			priors := NewTensor(g, Float64, tcase.input.Dims(), WithShape(tcase.input.Shape()...), WithInit(Ones()))
			x, err := tn.FeatureTransformer(FeatureTransformerOpts{
				VirtualBatchSize:  tcase.vbs,
				OutputDimension:   tcase.output,
				Shared:            nil,
				IndependentBlocks: tcase.independentBlocks,
				WeightsInit:       initDummyWeights,
			})(input, priors)

			if tcase.expectedErr != "" {
				c.Error(err)

				c.Equal(tcase.expectedErr, err.Error())

				return
			} else {
				c.NoError(err)
			}

			vm := NewTapeMachine(g)
			c.NoError(vm.RunAll())

			c.Equal(tcase.expectedShape, x.Shape())
			c.Equal(tcase.expectedOutput, x.Value().Data().([]float64))
		})
	}
}
