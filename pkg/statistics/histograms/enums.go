package histograms

type HistoIx int

const (
	HistoAlpha HistoIx = iota
	HistoAlphaPred
	HistoGreen
	HistoGreenPred
	HistoRed
	HistoRedPred
	HistoBlue
	HistoBluePred
	HistoRedSubGreen
	HistoRedPredSubGreen
	HistoBlueSubGreen
	HistoBluePredSubGreen
	HistoPalette
	HistoTotal // Must be last.
)
