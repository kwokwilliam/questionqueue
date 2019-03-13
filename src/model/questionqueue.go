package model

// QuestionQueue will be unmarshalled from the redis store
// this requires the json the queue receives to be in the format:
// {
//		"queue": [
// 			{ format of question struct }
//		]
// }
type QuestionQueue struct {
	Queue []*Question `json:"queue"`
}

// PositionInLine is the position in line for the student map
type PositionInLine struct {
	Position    int `json:"position"`
	QueueLength int `json:"queueLength"`
}

// GetStudentPositions will convert the entire queue into a map to get
// student positions faster.
func (q *QuestionQueue) GetStudentPositions() map[string]*PositionInLine {
	studentPositions := make(map[string]*PositionInLine)
	for i, question := range q.Queue {
		studentPositions[question.ID] = &PositionInLine{i + 1, len(q.Queue)}
	}
	return studentPositions
}
