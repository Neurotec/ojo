# ojo
!!!PROOF CONCEPT

Remote observer pattern.

This it's a single library can notify actions from remote observables to
central host/observer.

With aditional point, that the central host can feedback messages to observables.

## See our client

~~~go

type CallFieldUpdate struct {
	ID  string
	State string
}

type Call struct {
	...
}

// Receive a feedback from central host
func (call Call) Feedback(msg interface{}) {
	switch msg.(type) {
		case CallHangup:
		// do hangup a call
		
		// central goes down
		case ojo.Down:
	}
}

....

observer, err := ojo.Dial("localhost:9353")
if err != nil {
	log.Fatal(err)
}

fcall, err := observer.Observe(call)
....

fcall.Notify(CallFieldUpdate{call.ID, "answered"})
....
fcall.Notify(CallFieldUpdate{call.ID, "hangup"})
...
fcall.Close()

~~~

## See our server

~~~go

type CallManager struct {
	calls[string] Call
}

type (cm CallManager) OnNotify(observable interface{}, data chan interface{}, feedback ojo.Feedback) {
	for msg := range data {

		switch msg.(type) {
			case CallFieldUpdate:
			//update fields
			case ojo.ObservableClose:
			//we receive a close connection
		}
	}
	

}

....


ojo.ListenAndServer("localhost:9353", CallManager{})
~~~