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

// GetStudentizedQueue will return a copy of the original question queue with
// all data stripped except the ID of the individuals within the queue.
func (q *QuestionQueue) GetStudentizedQueue() *QuestionQueue {
	studentizedQueue := &QuestionQueue{}
	for _, question := range q.Queue {
		studentizedQueue.Queue = append(studentizedQueue.Queue, &Question{
			ID: question.ID,
		})
	}
	return studentizedQueue
}
