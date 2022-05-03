package datastructures

type UnboundedChanString struct {
	In     chan<- string // channel for write
	Out    <-chan string // channel for read
	buffer []string      // buffer
}

// Len returns len of Out plus len of buffer.
func (c UnboundedChanString) Len() int {
	return len(c.buffer) + len(c.Out)
}

// BufLen returns len of the buffer.
func (c UnboundedChanString) BufLen() int {
	return len(c.buffer)
}
func NewUnboundedChan(initCapacity int) UnboundedChanString {
	// Create chan type with three fields and unlimited cache
	in := make(chan string, initCapacity)
	out := make(chan string, initCapacity)
	ch := UnboundedChanString{In: in, Out: out, buffer: make([]string, 0, initCapacity)}
	// Through a goroutine, the data is continuously read out from in and put into out or buffer
	go func() {
		defer close(out) // in is closed, and out is also closed after the data is read
	loop:
		for {
			val, ok := <-in
			if !ok { // If in has been closed, exit loop
				break loop
			}
			// Otherwise try to put the data read from in into out
			select {
			case out <- val: //The success of putting in means that out is not full just now and there is no additional data in the buffer to be processed, so go back to the beginning of loop
				continue
			default:
			}
			// If out is full, you need to put the data into the cache
			ch.buffer = append(ch.buffer, val)
			// Process the cache and keep trying to put the data in the cache into the out, until there is no more data in the cache,
			// In order to avoid blocking the in channel, we also try to read data from in, because out is full at this time, so we put the data directly into the cache
			for len(ch.buffer) > 0 {
				select {
				case val, ok := <-in: // Read data from in, put it into the cache, if in is closed, exit loop
					if !ok {
						break loop
					}
					ch.buffer = append(ch.buffer, val)
				case out <- ch.buffer[0]: // Put the oldest data in the cache into out and move out the first element
					ch.buffer = ch.buffer[1:]
					if len(ch.buffer) == 0 { // Avoid memory leaks. If the cache has finished processing, revert to the original state
						ch.buffer = make([]string, 0, initCapacity)
					}
				}
			}
		}
		// After in is closed and the loop is exited, there may still be unprocessed data in the buffer, which needs to be stuffed into out
		// This logic is called "drain".
		// After this piece of logic is processed, the out can be closed out
		for len(ch.buffer) > 0 {
			out <- ch.buffer[0]
			ch.buffer = ch.buffer[1:]
		}
	}()
	return ch
}
