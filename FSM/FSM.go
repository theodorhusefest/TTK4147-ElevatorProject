package FSM

import (
	"../IO"
  "fmt"
  //"../Utilities"
	. "../Config"
  "time"

)


type FSMchannels struct{
  NewLocalOrderChan chan int
  ArrivedAtFloorChan chan int
  DoorTimeoutChan chan bool
}

func StateMachine(FSMchans FSMchannels, LocalOrderFinishedChan chan int, elevatorMatrix [][]int){
	elevator := Elevator{
			State: IDLE,
			Floor: elevatorMatrix[2][ElevID*NumElevators],
			Dir: DIR_Stop,
	}

	doorOpenTimeOut := time.NewTimer(3 * time.Second)
	doorOpenTimeOut.Stop()


  for {
    select {
    case newLocalOrder := <- FSMchans.NewLocalOrderChan:

			switch elevator.State {
			case IDLE:
				elevator.Dir = chooseDirection(elevatorMatrix, elevator)
				io.SetMotorDirection(elevator.Dir)
				if elevator.Dir == DIR_Stop {
					// Open door for 3 seconds
					io.SetDoorOpenLamp(true)
					//clearFloors(elevator,elevatorMatrix)
					elevator.State = DOOROPEN
					doorOpenTimeOut.Reset(3 * time.Second)
				}
				elevator.State = MOVING


			case MOVING:


			case DOOROPEN:
				if elevator.Floor == newLocalOrder {
					doorOpenTimeOut.Reset(3 * time.Second)
					fmt.Println("Resetting Time")
				}
			}


    case currentFloor := <- FSMchans.ArrivedAtFloorChan:
        //Update floor in matrix
        updateElevFloor(0,currentFloor,elevatorMatrix)
				elevator.Floor = currentFloor
				io.SetFloorIndicator(currentFloor)
        if shouldStop(0, elevator, elevatorMatrix) {
					elevator.State = DOOROPEN
					io.SetDoorOpenLamp(true)
					doorOpenTimeOut.Reset(3 * time.Second)
					//clearFloors(elevator, elevatorMatrix)
          fmt.Println("stop")
					io.SetMotorDirection(DIR_Stop)
        }


		case <-doorOpenTimeOut.C:
			io.SetDoorOpenLamp(false)
			//clearFloors(elevator, elevatorMatrix)
			LocalOrderFinishedChan <- elevator.Floor
			elevator.Dir = chooseDirection(elevatorMatrix, elevator)
			if elevator.Dir == DIR_Stop {
				elevator.State = IDLE
			} else {
				io.SetMotorDirection(elevator.Dir)
				elevator.State = MOVING
			}

    }
  }
}

func matrixIsEmpty(elevatorMatrix [][]int) bool{
	for floor := 4; floor < 4+ NumFloors; floor++ {
    for buttons := (ElevID*3); buttons < (ElevID*3 + 3); buttons++ {
      if elevatorMatrix[floor][buttons] == 1 {
        return false
      }
    }
  }
	return true
}

func updateElevFloor(ElevID int, newFloor int, elevatorMatrix [][]int){
  elevatorMatrix[2][ElevID*3] = newFloor
  //tror man ikke trenger å returnere matriser
}

func isOrderAbove(ElevID int, currentFloor int, elevatorMatrix [][]int) bool {
	if currentFloor == 3 {
		return false
	}

	for floor := (len(elevatorMatrix) - currentFloor - 2); floor > 3; floor-- {
    for buttons := (ElevID*3); buttons < (ElevID*3 + 3); buttons++ {
      if elevatorMatrix[floor][buttons] == 1 {
        return true
      }
    }
  }
  return false
}


func isOrderBelow(ElevID int, currentFloor int, elevatorMatrix [][]int) bool {
	if currentFloor == 0 {
		return false
	}

	for floor := (len(elevatorMatrix) - currentFloor); floor < (len(elevatorMatrix)); floor++ {
    for buttons := (ElevID*3); buttons < (ElevID*3 + 3); buttons++ {
      if elevatorMatrix[floor][buttons] == 1 {
        return true
      }
    }

  }
  return false
}

func shouldStop(ElevID int, elevator Elevator, elevatorMatrix [][]int) bool {
  //Cab call is pressed, stop
  if elevatorMatrix[len(elevatorMatrix) - elevator.Floor - 1][ElevID*NumElevators + 2] == 1 {
    return true
  }
	switch elevator.Dir{
	case DIR_Up:
		if elevatorMatrix[len(elevatorMatrix) - elevator.Floor - 1][ElevID*NumElevators] == 1 {
	    return true
	  } else if elevatorMatrix[len(elevatorMatrix) - elevator.Floor - 1][ElevID*NumElevators + 1] == 1 && !isOrderAbove(ElevID, elevator.Floor, elevatorMatrix) {
			return true
		}
	case DIR_Down:
		if elevatorMatrix[len(elevatorMatrix) - elevator.Floor - 1][ElevID*NumElevators + 1] == 1 {
	    return true
	  } else if elevatorMatrix[len(elevatorMatrix) - elevator.Floor - 1][ElevID*NumElevators] == 1 && !isOrderBelow(ElevID, elevator.Floor, elevatorMatrix) {
			return true
		}
	}
  return false
}

func chooseDirection(elevatorMatrix [][]int, elevator Elevator) MotorDirection {

	if isOrderAbove(ElevID, elevator.Floor, elevatorMatrix){
		return DIR_Up
	}
	if isOrderBelow(ElevID, elevator.Floor, elevatorMatrix){
		return DIR_Down
	}
	return DIR_Stop
}
