module whisper/services/user

go 1.22.0

require (
	whisper/libs/domain v0.0.0
	whisper/libs/shared v0.0.0
)

replace whisper/libs/domain => ../../libs/domain
replace whisper/libs/shared => ../../libs/shared
