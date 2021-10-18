package tabnet

import (
	"testing"

	"github.com/dcu/godl"
	"github.com/stretchr/testify/require"
	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func TestFeatureTransformer(t *testing.T) {
	testCases := []struct {
		desc              string
		input             tensor.Tensor
		vbs               int
		independentBlocks int
		sharedBlocks      int
		output            int
		expectedShape     tensor.Shape
		expectedErr       string
		expectedOutput    []float64
		expectedGrad      []float64
		expectedCost      float64
	}{
		{
			desc: "Example1",
			input: tensor.New(
				tensor.WithShape(6, 2),
				tensor.WithBacking([]float64{0.4, 1.4, 2.4, 3.4, 4.4, 5.4, 6.4, 7.4, 8.4, 9.4, 10.4, 11.4}),
			),
			vbs:               2,
			output:            2,
			independentBlocks: 5,
			sharedBlocks:      0,
			expectedShape:     tensor.Shape{6, 2},
			expectedOutput:    []float64{-0.46379327241699114, -0.28701657712035417, 1.87719497530188, 2.053971670598517, 0.24331350876955665, 0.4200902040661938, 2.5843017564884274, 2.761078451785065, 0.9504202899561044, 1.1271969852527415, 3.291408537674976, 3.4681852329716127},
			expectedGrad:      []float64{0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333, 0.08333333333333333},
			expectedCost:      1.502195980277311,
		},
		{
			desc: "Example2",
			input: tensor.New(
				tensor.WithShape(6, 2),
				tensor.WithBacking([]float64{0.4, 1.4, 2.4, 3.4, 4.4, 5.4, 6.4, 7.4, 8.4, 9.4, 10.4, 11.4}),
			),
			vbs:               2,
			output:            16,
			independentBlocks: 5,
			sharedBlocks:      5,
			expectedShape:     tensor.Shape{6, 16},
			expectedOutput:    []float64{-0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, -0.6324731811198426, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826, 1.719240284475826},
			expectedGrad:      []float64{0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666, 0.010416666666666666},
			expectedCost:      0.5433835516779918,
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.desc, func(t *testing.T) {
			c := require.New(t)

			tn := godl.NewModel()
			tn.Training = true

			g := tn.ExprGraph()

			input := gorgonia.NewTensor(g, tensor.Float64, tcase.input.Dims(), gorgonia.WithShape(tcase.input.Shape()...), gorgonia.WithName("input"+tcase.desc), gorgonia.WithValue(tcase.input))

			shared := make([]godl.Layer, tcase.sharedBlocks)
			fcInput := input.Shape()[1]
			fcOutput := 2 * (8 + 8)
			for i := 0; i < tcase.sharedBlocks; i++ {
				shared[i] = godl.FC(tn, godl.FCOpts{
					OutputDimension: fcOutput, // double the size so we can take half and half
					WeightsInit:     initDummyWeights,
					InputDimension:  fcInput,
				})

				fcInput = 8 + 8
			}

			y, err := FeatureTransformer(tn, FeatureTransformerOpts{
				VirtualBatchSize:  tcase.vbs,
				InputDimension:    input.Shape()[1],
				OutputDimension:   tcase.output,
				Shared:            shared,
				IndependentBlocks: tcase.independentBlocks,
				WeightsInit:       initDummyWeights,
			})(input)

			if tcase.expectedErr != "" {
				c.Error(err)

				c.Equal(tcase.expectedErr, err.Error())

				return
			} else {
				c.NoError(err)
			}

			cost := gorgonia.Must(gorgonia.Mean(y.Output))
			_, err = gorgonia.Grad(cost, input)
			c.NoError(err)

			vm := gorgonia.NewTapeMachine(g, gorgonia.BindDualValues(tn.Learnables()...))
			c.NoError(vm.RunAll())

			tn.PrintWatchables()

			c.Equal(tcase.expectedShape, y.Shape())
			c.Equal(tcase.expectedOutput, y.Value().Data().([]float64))

			yGrad, err := y.Output.Grad()
			c.NoError(err)

			c.Equal(tcase.expectedGrad, yGrad.Data())
			c.Equal(tcase.expectedCost, cost.Value().Data())
		})
	}
}
