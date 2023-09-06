package server

// Option represents an server option.
type Option func(*Server)

// WithAddr sets the listen for the server.
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithTempLocation sets the location where image files should be uploaded to.
func WithTempLocation(tmp string) Option {
	return func(s *Server) {
		s.tmpLocation = tmp
	}
}

// WithChunkSize sets the chunkSize supported by the server
func WithChunkSize(chunkSize int) Option {
	return func(s *Server) {
		s.chunkSize = chunkSize
	}
}
