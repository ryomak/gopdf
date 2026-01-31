package layout

import "fmt"

// AdjustLayout はPageLayoutを自動調整する
func (pl *PageLayout) AdjustLayout(opts LayoutAdjustmentOptions) error {
	switch opts.Strategy {
	case StrategyFlowDown:
		return pl.adjustLayoutFlowDown(opts)
	case StrategyCompact:
		return pl.adjustLayoutCompact(opts)
	case StrategyEvenSpacing:
		return pl.adjustLayoutEvenSpacing(opts)
	case StrategyFitContent:
		return pl.adjustLayoutFitContent(opts)
	case StrategyPreservePosition:
		// 位置を保持するので何もしない
		return nil
	default:
		return fmt.Errorf("unsupported layout strategy: %s", opts.Strategy)
	}
}
