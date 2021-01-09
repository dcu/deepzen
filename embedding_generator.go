package tabnet

import (
	"sort"

	"gorgonia.org/gorgonia"
	"gorgonia.org/tensor"
)

func (m *Model) EmbeddingGenerator(inputDims int, catDims []int, catIdxs []int, catEmbDim []int, opts EmbeddingOpts) Layer {
	skipEmbedding := false
	if len(catDims) == 0 || len(catIdxs) == 0 {
		skipEmbedding = true
	}

	sort.Slice(catIdxs, func(i, j int) bool {
		return catIdxs[i] < catIdxs[j]
	})

	embeddings := make([]Layer, len(catIdxs))
	continuousIdxs := make([]bool, inputDims) // FIXME: name

	for i, v := range catIdxs {
		embeddings[i] =
			m.Embedding(
				catDims[i],
				catEmbDim[i],
				opts,
			)

		continuousIdxs[v] = true
	}

	return func(inputs ...*gorgonia.Node) (*gorgonia.Node, error) {
		err := m.checkArity("EmbeddingGenerator", inputs, 1)
		if err != nil {
			return nil, err
		}

		x := inputs[0]
		if skipEmbedding {
			return x, nil
		}

		cols := make([]*gorgonia.Node, len(continuousIdxs))
		catFeatCounter := 0

		for featInitIdx, isContinuous := range continuousIdxs {
			s := gorgonia.Must(gorgonia.Slice(x, nil, gorgonia.S(featInitIdx)))

			if !isContinuous {
				cols[featInitIdx] = gorgonia.Must(gorgonia.Reshape(s, tensor.Shape{s.Shape().TotalSize(), 1}))
			} else {
				cols[featInitIdx] = gorgonia.Must(embeddings[catFeatCounter](s))

				catFeatCounter++
			}
		}

		result := gorgonia.Must(gorgonia.Concat(1, cols...))

		return result, nil
	}
}