package tabnet

import (
	"fmt"

	"gorgonia.org/gorgonia"
)

// GLU implements a Gated Linear Unit Block
func (nn *TabNet) GLU(opts GLUOpts) Layer {
	if opts.ActivationFn == nil {
		opts.ActivationFn = gorgonia.Sigmoid
	}

	return func(x *gorgonia.Node) (*gorgonia.Node, error) {
		var (
			fc  *gorgonia.Node
			err error
		)

		if opts.FC == nil {
			opts.FC = nn.FC(FCOpts{
				OutputFeatures: opts.OutputFeatures * 2,
			})
		}

		fc, err = opts.FC(x)
		if err != nil {
			return nil, fmt.Errorf("[glu] applying FC failed: %w", err)
		}

		gbn, err := nn.GBN(fc, GBNOpts{
			VirtualBatchSize: opts.VirtualBatchSize,
		})
		if err != nil {
			return nil, fmt.Errorf("[glu] applying GBN failed: %w", err)
		}

		// GLU
		firstHalf := gorgonia.Must(gorgonia.Slice(gbn, gorgonia.S(0, opts.OutputFeatures)))
		secondHalf := gorgonia.Must(gorgonia.Slice(gbn, gorgonia.S(opts.OutputFeatures, gbn.Shape()[1])))

		act, err := opts.ActivationFn(secondHalf)
		if err != nil {
			return nil, fmt.Errorf("[glu] applying activation function failed: %w", err)
		}

		mul, err := gorgonia.HadamardProd(firstHalf, act)
		if err != nil {
			return nil, fmt.Errorf("[glu] HadamardProd %d x %d: %w", firstHalf.Shape(), act.Shape(), err)
		}

		return mul, nil
	}
}