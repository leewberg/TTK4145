Code for project in TTK4145 Real time programming, group 5
======================

Soltion is centred around a collective database which is consistent accross all nodes, and the nodes make actions based on the database.

To run the program, run 

`elevatorserver --port [15657 + elevatorID]`

Then, open another terminal window and run

`go run main.go [elevatorID]`

`elevatorID` is zero-indexed and specifies the ID of the elevator. For example, if you want to test the system with three elevators, `elevatorID` is a integer 0-2.
