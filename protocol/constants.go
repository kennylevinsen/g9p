package protocol

const (
	HeaderSize = 4 + 1
)

// MessageType constants
const (
	Tversion MessageType = 100 + iota
	Rversion
	Tauth
	Rauth
	Tattach
	Rattach
	Terror
	Rerror
	Tflush
	Rflush
	Twalk
	Rwalk
	Topen
	Ropen
	Tcreate
	Rcreate
	Tread
	Rread
	Twrite
	Rwrite
	Tclunk
	Rclunk
	Tremove
	Rremove
	Tstat
	Rstat
	Twstat
	Rwstat
	Tlast
)

// Special message values
const (
	NOTAG Tag = 0xFFFF
	NOFID Fid = 0xFFFFFFFF
)

// Opening modes
const (
	OREAD OpenMode = iota
	OWRITE
	ORDWR
	OEXEC

	OTRUNC OpenMode = 16 * (iota + 1)
	OCEXEC
	ORCLOSE
)

// Permission bits
const (
	DMDIR       FileMode = 0x80000000
	DMAPPEND    FileMode = 0x40000000
	DMEXCL      FileMode = 0x20000000
	DMMOUNT     FileMode = 0x10000000
	DMAUTH      FileMode = 0x08000000
	DMTMP       FileMode = 0x04000000
	DMSYMLINK   FileMode = 0x02000000
	DMLINK      FileMode = 0x01000000
	DMDEVICE    FileMode = 0x00800000
	DMNAMEDPIPE FileMode = 0x00200000
	DMSOCKET    FileMode = 0x00100000
	DMSETUID    FileMode = 0x00080000
	DMSETGID    FileMode = 0x00040000
	DMREAD      FileMode = 0x4
	DMWRITE     FileMode = 0x2
	DMEXEC      FileMode = 0x1
)

// Qid types
const (
	QTFILE    QidType = 0x00
	QTLINK    QidType = 0x01
	QTSYMLINK QidType = 0x02
	QTTMP     QidType = 0x04
	QTAUTH    QidType = 0x08
	QTMOUNT   QidType = 0x10
	QTEXCL    QidType = 0x20
	QTAPPEND  QidType = 0x40
	QTDIR     QidType = 0x80
)
