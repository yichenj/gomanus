package llm

type ModelType string

const (
	ModelTypeLanguageModel   ModelType = "language"
	ModelTypeReasonModel     ModelType = "reason"
	ModelTypeMultiModalModel ModelType = "multi-modal"
)

type TypedModelSettings map[ModelType]ModelSetting

type ModelSetting struct {
	ModelName string
	BaseURL   string
	APIKey    string
}
