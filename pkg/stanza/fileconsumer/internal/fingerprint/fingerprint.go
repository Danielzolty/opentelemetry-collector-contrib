package fingerprint

type Fingerprint interface {
	Equal(other Fingerprint) bool

	StartsWith(old Fingerprint) bool

	Copy() *Fingerprint

	ByteSize() int

	IsMaxSize(maxFingerprintSize int, offset int64) bool
}
