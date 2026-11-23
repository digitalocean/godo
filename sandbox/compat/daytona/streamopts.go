package daytona

// downloadStreamConfig and uploadStreamConfig hold the progress callbacks that
// WithDownloadProgress and WithUploadProgress set. The stream methods that
// read them are not implemented yet, so nothing consumes the values.
type downloadStreamConfig struct{ progress func(DownloadProgress) }

type uploadStreamConfig struct{ progress func(UploadProgress) }

// WithDownloadProgress mirrors daytona.WithDownloadProgress.
func WithDownloadProgress(fn func(DownloadProgress)) DownloadStreamOption {
	return func(c *downloadStreamConfig) { c.progress = fn }
}

// WithUploadProgress mirrors daytona.WithUploadProgress.
func WithUploadProgress(fn func(UploadProgress)) UploadStreamOption {
	return func(c *uploadStreamConfig) { c.progress = fn }
}
