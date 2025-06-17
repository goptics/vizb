package main

import (
	"errors"
	"testing"

	"github.com/goptics/varmq"
)

func resultTask(j varmq.Job[int]) (int, error) {
	return j.Data() * 2, nil
}

func task(j varmq.Job[int]) {
	_ = j.Data() * 2
}

func errTask(j varmq.Job[int]) error {
	// Simple task that returns error for odd numbers
	if j.Data()%2 != 0 {
		return errors.New("odd number error")
	}
	return nil
}

// BenchmarkAdd benchmarks the Add operation across different queue types.
func BenchmarkAdd(b *testing.B) {
	b.Run("Queue", func(b *testing.B) {
		// Create a worker with the double function
		worker := varmq.NewWorker(task)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for j := range b.N {
			job, _ := q.Add(j)
			job.Wait()
		}
	})

	b.Run("Priority", func(b *testing.B) {
		// Create a worker with the double function
		worker := varmq.NewWorker(task)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for i := range b.N {
			if job, ok := q.Add(i, i%10); ok {
				job.Wait()
			}
		}
	})

	b.Run("ErrQueue", func(b *testing.B) {
		// Create an error worker
		worker := varmq.NewErrWorker(errTask)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for j := range b.N {
			if job, ok := q.Add(j); ok {
				job.Err()
			}
		}
	})

	b.Run("ErrPriority", func(b *testing.B) {
		// Create an error worker
		worker := varmq.NewErrWorker(errTask)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for i := range b.N {
			if job, ok := q.Add(i, i%10); ok {
				job.Err()
			}
		}
	})

	b.Run("ResultQueue", func(b *testing.B) {
		// Create a result worker
		worker := varmq.NewResultWorker(resultTask)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for j := range b.N {
			if job, ok := q.Add(j); ok {
				job.Result()
			}
		}
	})

	b.Run("ResultPriority", func(b *testing.B) {
		// Create a result worker
		worker := varmq.NewResultWorker(resultTask)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		b.ResetTimer()
		for i := range b.N {
			if job, ok := q.Add(i, i%10); ok {
				job.Result()
			}
		}
	})
}

// BenchmarkAddAll benchmarks the AddAll operation across different queue types.
func BenchmarkAddAll(b *testing.B) {
	b.Run("Queue", func(b *testing.B) {
		// Create a worker with the double function
		worker := varmq.NewWorker(task)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})

	b.Run("Priority", func(b *testing.B) {
		// Create a worker with the double function
		worker := varmq.NewWorker(task)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i, Priority: i % 10}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})

	b.Run("ErrQueue", func(b *testing.B) {
		// Create an error worker
		worker := varmq.NewErrWorker(errTask)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})

	b.Run("ErrPriority", func(b *testing.B) {
		// Create an error worker
		worker := varmq.NewErrWorker(errTask)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i, Priority: i % 10}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})

	b.Run("ResultQueue", func(b *testing.B) {
		// Create a result worker
		worker := varmq.NewResultWorker(resultTask)
		// Bind the worker to a standard queue
		q := worker.BindQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})

	b.Run("ResultPriority", func(b *testing.B) {
		// Create a result worker
		worker := varmq.NewResultWorker(resultTask)
		// Bind the worker to a priority queue
		q := worker.BindPriorityQueue()
		defer worker.WaitAndStop()

		data := make([]varmq.Item[int], 1000) // Using a constant size of 1000 for testing
		for i := range data {
			data[i] = varmq.Item[int]{Data: i, Priority: i % 10}
		}

		b.ResetTimer()
		for b.Loop() {
			q.AddAll(data).Wait()
		}
	})
}
