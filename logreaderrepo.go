package main

//LogReaderRepo is used to hold the list of available LogReaders
type LogReaderRepo struct {
	LogReaders map[LogType]LogReader
}

//NewLogReaderRepo returns an instance of the LogReader repository
func NewLogReaderRepo() LogReaderRepository {
	return &LogReaderRepo{
		LogReaders: map[LogType]LogReader{
			Transaction: NewTransactionLogReader(),
			Sumo:        NewSumoLogReader(),
		},
	}
}

//LogReaderRepository is the repository for registered log readers
type LogReaderRepository interface {
	GetLogReader(logType LogType) LogReader
}

//GetLogReader returns the log reader for the specified type
func (lrr *LogReaderRepo) GetLogReader(logType LogType) LogReader {
	lr := lrr.LogReaders[logType]
	return lr
}
