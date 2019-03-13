package store

// Question is used for individual questions
type Question struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Class     string `json:"class,omitempty"`
	Topic     string `json:"topic,omitempty"`
	Problem   string `json:"problem,omitempty"`
	LocationX string `json:"loc.x,omitempty"`
	LocationY string `json:"loc.y,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

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
	Position int `json:"position"`
}

// GetStudentPositions will convert the entire queue into a map to get
// student positions faster.
func (q *QuestionQueue) GetStudentPositions() map[string]*PositionInLine {
	studentPositions := make(map[string]*PositionInLine)
	for i, question := range q.Queue {
		studentPositions[question.ID] = &PositionInLine{i}
	}
	return studentPositions
}
