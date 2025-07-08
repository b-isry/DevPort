package types

type Manifest struct {
	Files map[string]string `json:"files"`
	NativeModulePaths []string `json:"nativeModulePaths"`
}
