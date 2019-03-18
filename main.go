package main

import (
//  "fmt"
  . "./Config"
  "./Initialize"
//  "./Utilities"
  "./orderManager"
  "./IO"
  "./FSM"
  "./elevatorSync"
  "./Network/network/peers"
  "./Network/network/bcast"
  "time"
  "strconv"

)


func main() {

  // Initialize
  // !!!!!!!!!!!!!! Skru av alle lys, initialiser matrisen til antall input
  elevatorMatrix, elevConfig := initialize.Initialize()

  io.Init("localhost:15657",4)

  // Channels for FSM
  FSMchans := FSM.FSMchannels{
    NewLocalOrderChan: make(chan int),
    ArrivedAtFloorChan: make(chan int),
    DoorTimeoutChan:  make(chan bool),
  }
  // Channels for OrderManager
  OrderManagerchans := orderManager.OrderManagerChannels{
    UpdateElevatorChan: make(chan Message),
    LocalOrderFinishedChan: make(chan int),
  }
  // Channels for SyncElevator
  SyncElevatorChans := syncElevator.SyncElevatorChannels{
    OutGoingMsg: make(chan Message),
    InCommingMsg: make(chan Message),
    ChangeInOrderch: make(chan Message),
    PeerUpdate: make(chan peers.PeerUpdate),
    TransmitEnable: make(chan bool),
    BroadcastTicker: make(chan bool),
  }
  var (
    NewGlobalOrderChan = make(chan ButtonEvent)
  )

  channelFloor := make(chan int) //channel that is used in InitElevator. Should maybe have a struct with channels?
  //elevatorMatrix := initialize.InitializeMatrix(NumFloors,NumElevators)  // Set up matrix, add ID
  initialize.InitElevator(elevConfig,elevatorMatrix,channelFloor)  // Move elevator to nearest floor and update matrix
//  utilities.PrintMatrix(elevatorMatrix, elevConfig.NumFloors,elevConfig.NumElevators)

  // Goroutines used in FSM
  go io.PollFloorSensor(FSMchans.ArrivedAtFloorChan)
  go FSM.StateMachine(FSMchans, OrderManagerchans.LocalOrderFinishedChan, elevatorMatrix,elevConfig)

  // Goroutines used in OrderManager
  go io.PollButtons(NewGlobalOrderChan)
  go orderManager.OrderManager(OrderManagerchans, NewGlobalOrderChan, FSMchans.NewLocalOrderChan, elevatorMatrix, SyncElevatorChans.OutGoingMsg, SyncElevatorChans.ChangeInOrderch, elevConfig)

  // Goroutines used in SyncElevator
  go syncElevator.SyncElevator(SyncElevatorChans, elevConfig, OrderManagerchans.UpdateElevatorChan)

  // Goroutines used in Network/Peers
  go peers.Transmitter(15789, strconv.Itoa(elevConfig.ElevID), SyncElevatorChans.TransmitEnable)
  go peers.Receiver(15789, SyncElevatorChans.PeerUpdate)

  //  Goroutines used in Network/Bcast
  go bcast.Transmitter(15790, SyncElevatorChans.OutGoingMsg)
  go bcast.Receiver(15790, SyncElevatorChans.InCommingMsg)


  time.Sleep(10*time.Second)

  select{}
}
