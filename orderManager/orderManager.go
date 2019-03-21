package orderManager

import(
  . "../Config"
  //"../Utilities"
  "../IO"
  "../Config"
  "../hallRequestAssigner"
  "fmt"
  "time"
)

/*
Matrix  ID_1        -----------   -----     ID_2        -----------   -----     ID_3        -----------   -----
        State_1     -----------   -----     State_2     -----------   -----     State_3     -----------   -----
        FLoor_1     -----------   -----     FLoor_2     -----------   -----     FLoor_3     -----------   -----
        Dir_1       -----------   -----     Dir_2       -----------   -----     Dir_3       -----------   -----
        --------    Down_floor_4  Cab_4     --------    Down_floor_4  Cab_4     --------    Down_floor_4  Cab_4
        Up_floor_3  Down_floor 3  Cab_3     Up_floor_3  Down_floor 3  Cab_3     Up_floor_3  Down_floor 3  Cab_3
        Up_floor_2  Down_floor_2  Cab_2     Up_floor_2  Down_floor_2  Cab_2     Up_floor_2  Down_floor_2  Cab_2
        Up_floor_1  --------      Cab_1     Up_floor_1  --------      Cab_1     Up_floor_1  --------      Cab_1
*/



type OrderManagerChannels struct{
  LocalOrderFinishedChan chan int
  NewLocalOrderch chan Message
  UpdateOrderch chan Message
  MatrixUpdatech chan Message

}


func OrderManager(elevatorMatrix [][]int, elevatorConfig config.ElevConfig, OrderManagerChans OrderManagerChannels, 
                  NewGlobalOrderChan chan ButtonEvent, NewLocalOrderChan chan int,
                  OutGoingMsg chan []Message, ChangeInOrderch chan []Message, UpdateElevStatusch chan Message) {

  GlobalOrderTimedOut := time.NewTimer(5 * time.Second)
  GlobalOrderTimedOut.Stop()

  for {
      select {

        /*
        1: Ordre tas imot av en heis.
        2: Den heisen kjører kostfunksjon og bestemmer hvem som får jobben.
        3: Heisen sender ordren til alle andre heiser, så alle er oppdatert.
        4: Alle skrur
        5: Den heisen som får jobben, trigger sin egen FSM med        NewLocalOrderChan <- int(newGlobalOrder.Floor)
        6: Heisen som har utført et oppdrag oppdaterer de andre, så alle vet det samme
        7: Alle fjerner ordren lokalt og evt. skrur av lys
        */

      // -----------------------------------------------------------------------------------------------------Case triggered by local button
      case NewGlobalOrder := <- NewGlobalOrderChan:

        switch NewGlobalOrder.Button {

        case BT_Cab:

        outMessage := []Message {{Select: NewOrder, Done: false, ID: elevatorConfig.ElevID, Floor: NewGlobalOrder.Floor, Button: NewGlobalOrder.Button}}
        // Send message to sync

        ChangeInOrderch <- outMessage

       // addOrder(elevatorConfig.ElevID, elevatorMatrix, NewGlobalOrder) // ELEV-ID TO DEDICATED ELEVATOR
      //  setLight(outMessage[0], elevatorConfig)


        default:

        newHallOrders := hallOrderAssigner.AssignHallOrder(NewGlobalOrder, elevatorMatrix)
        fmt.Println()



        // Send message to sync                //time.Sleep(10*time.Second)

        ChangeInOrderch <- newHallOrders
        /*
        // Wait for sync to say everyone knows the same
  */
      }

      case OrderUpdate := <- OrderManagerChans.UpdateOrderch:
          
          switch OrderUpdate.Select {
              case NewOrder:
                  localOrder := ButtonEvent{Floor: OrderUpdate.Floor, Button: OrderUpdate.Button}
                  addOrder(OrderUpdate.ID, elevatorMatrix, localOrder)
                  setLight(OrderUpdate, elevatorConfig)
                  if OrderUpdate.ID == elevatorConfig.ElevID {
                      NewLocalOrderChan <- OrderUpdate.Floor
                      fmt.Println("Sending LocalOrder")
                  }

              case OrderComplete:
                  clearFloors(OrderUpdate.Floor, elevatorMatrix, OrderUpdate.ID)
                  clearLight(OrderUpdate.Floor)

          }


      case StateUpdate := <- UpdateElevStatusch:

          switch StateUpdate.Select {
              case UpdateStates:

                  InsertID(StateUpdate.ID, elevatorMatrix)
                  InsertState(StateUpdate.ID, StateUpdate.State, elevatorMatrix)
                  InsertDirection(StateUpdate.ID, StateUpdate.Dir, elevatorMatrix)
                  InsertFloor(StateUpdate.ID, StateUpdate.Floor, elevatorMatrix)

              case UpdateOffline:
                  InsertState(StateUpdate.ID, int(OFFLINE), elevatorMatrix)

              }


      case MatrixUpdate := <- OrderManagerChans.MatrixUpdatech:

          switch MatrixUpdate.Select {
            case SendMatrix:
                fmt.Println("Someone new on network, sending matrix")
                outMessage := []Message{{Select: UpdatedMatrix, Matrix: elevatorMatrix, ID: MatrixUpdate.ID}}
                ChangeInOrderch <- outMessage


            case UpdatedMatrix:
                if MatrixUpdate.ID == elevatorConfig.ElevID{
                    fmt.Println("Resetting matrix")
                    elevatorMatrix = updateOrdersInMatrix(elevatorMatrix, MatrixUpdate.Matrix, MatrixUpdate.ID)
                }
            }





      // -----------------------------------------------------------------------------------------------------Case triggered by elevator done with order
      case LocalOrderFinished := <- OrderManagerChans.LocalOrderFinishedChan:

        // Update message to be sent to everyone. Select = 2 for order done

        outMessage := []Message {{Select: OrderComplete, Done: false, ID: elevatorConfig.ElevID, Floor: LocalOrderFinished}}

        // Send message to sync
        ChangeInOrderch <- outMessage

        // Wait for sync to say everyone knows the same

        // Print updated matrix for fun)


      // ------------------------------------------------------------------------------------------------------- Case triggered every 5 seconds to check if orders left
      case <- GlobalOrderTimedOut.C:
        fmt.Println("GlobalOrderTimedOut")
        GlobalOrderTimedOut.Reset(5 * time.Second)



      // -------------------------------------------------------------------------------------------------------Case triggered by incomming update (New_order, order_done etc.)
    }
  }
}


func ordersLeftInMatrix(elevatorMatrix [][]int, elevatorConfig ElevConfig) int {
  for floor := 4; floor < 4+ NumFloors; floor++ {
    for buttons := (elevatorConfig.ElevID*3); buttons < (elevatorConfig.ElevID*3 + 3); buttons++ {
      if elevatorMatrix[floor][buttons] == 1 {
        return floor
      }
    }
  }
  return -1
}


func updateOrdersInMatrix(newMatrix [][]int, oldMatrix [][]int, id int) [][]int {
    for i := 0; i < (4+NumFloors); i ++ {
        for j := 0; j < 3*NumElevators; j ++ {
            if j != 3*id || i > 3 {
                newMatrix[i][j] = oldMatrix[i][j]
            }

        }
    }
    return newMatrix
}


func UpdateElevStatus(elevatorMatrix [][]int, FSMUpdateElevStatusch chan Message, ChangeInOrderch chan []Message) {
    for {
        select {
        case message := <- FSMUpdateElevStatusch:
            InsertID(message.ID, elevatorMatrix)
            InsertState(message.ID, message.State, elevatorMatrix)
            InsertDirection(message.ID, message.Dir, elevatorMatrix)
            InsertFloor(message.ID, message.Floor, elevatorMatrix)

            OutMessage := []Message{message}
            ChangeInOrderch <- OutMessage
        }
    }
}





func addOrder(id int, matrix [][]int, buttonPressed ButtonEvent) [][]int{
  matrix[NumFloors+3-buttonPressed.Floor][id*NumElevators + int(buttonPressed.Button)] = 1
  return matrix
}


func clearFloors(currentFloor int, elevatorMatrix [][]int, id int) {
	for button:=0; button < NumElevators; button++ {
		elevatorMatrix[len(elevatorMatrix)-currentFloor-1][button+id*NumElevators] = 0
	}
}

func InsertID(id int, matrix [][]int) {
    matrix[0][id*3] = id
}


func InsertFloor(id int, newFloor int, matrix [][]int){
  matrix[2][id*3] = newFloor
}

func InsertState(id int, state int, matrix [][]int){
  matrix[1][3*id] = state
}

func InsertDirection(id int, dir MotorDirection, matrix [][]int){
  switch dir{
    case DIR_Up:
      matrix[3][3*id] = 1
    case DIR_Down:
      matrix[3][3*id] = -1
    case DIR_Stop:
      matrix[3][3*id] = 0
  }
}



func setLight(illuminateOrder Message, elevatorConfig config.ElevConfig) {
  if illuminateOrder.ID == elevatorConfig.ElevID {
    io.SetButtonLamp(illuminateOrder.Button, illuminateOrder.Floor, true)
  } else if int(illuminateOrder.Button) == 2 {

  } else {
    io.SetButtonLamp(illuminateOrder.Button, illuminateOrder.Floor, true)
  }
}


func clearLight(LocalOrderFinished int) {
	io.SetButtonLamp(BT_Cab, LocalOrderFinished, false)
  io.SetButtonLamp(BT_HallUp, LocalOrderFinished, false)
  io.SetButtonLamp(BT_HallDown, LocalOrderFinished, false)
}




// FIKS
// Trigg fsm hvis ordre ikke blir gjort
// Button cab light
// initialize lights
// Ack
// Bug med 2 knapper samtidig
//
