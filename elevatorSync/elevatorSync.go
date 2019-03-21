package syncElevator

import (
  "fmt"
  "time"
  "strconv"
  "../Network/network/peers"
  . "../Config"

)



type SyncElevatorChannels struct{
  OutGoingMsg chan []Message
  InCommingMsg chan []Message
  ChangeInOrderch chan []Message
  PeerUpdate chan peers.PeerUpdate
  TransmitEnable chan bool
  BroadcastTicker chan bool
}

func SyncElevator(elevatorMatrix [][]int, syncChans SyncElevatorChannels, elevatorConfig ElevConfig, 
                  UpdateOrderch chan Message, UpdateElevStatusch chan Message, MatrixUpdatech chan Message){

  Online := false

//  broadcastTicker(syncChans)

  broadCastTicker := time.NewTicker(100 * time.Millisecond)
  //online := false
  for{
    select {

    // --------------------------------------------------------------------------Case triggered by local ordermanager, change in order
    case changeInOrder := <- syncChans.ChangeInOrderch:
      //Håndter endring som kom fra ordermanager. Send alt inn på message og sett message.Done = false
      //message := changeInOrder

        switch Online {
            case true: 
                select {
      
                    case <- broadCastTicker.C:
                        fmt.Println("Sending Message")
                        syncChans.OutGoingMsg <- changeInOrder
                    }

            case false:
                fmt.Println("Recieved order offline")
                for _, message := range changeInOrder {
                    fmt.Println("Type of message:", message.Select)
                    if !(message.Done) {
                      //SELECT = 1: NEW ORDER

                      switch  message.Select {

                      case NewOrder:
                          UpdateOrderch <- message

                      case OrderComplete:
                          UpdateOrderch <- message

                      case UpdateStates:
                          UpdateElevStatusch <- message

                      case UpdateOffline:
                          UpdateElevStatusch <- message

                      case ACK:
                          // ACKNOWLEDGE

                      case SendMatrix:
                          MatrixUpdatech <- message


                      case UpdatedMatrix:
                          MatrixUpdatech <- message
                      
                      message.Done = true
                    }
                  }
                }
              }


      // Broadcast message


      // Vent til alle er enige, gi klarsignal til ordermanager ??????

      // Sett message.Done = true



    // --------------------------------------------------------------------------Case triggered by bcast.Recieving
    case msgRecieved := <- syncChans.InCommingMsg:
      fmt.Println("Recieving Msg")
      for _, message := range msgRecieved {
          if !(message.Done) {
            //SELECT = 1: NEW ORDER

            switch  message.Select {

            case NewOrder:
                UpdateOrderch <- message

            case OrderComplete:
                UpdateOrderch <- message

            case UpdateStates:
                UpdateElevStatusch <- message

            case UpdateOffline:
                UpdateElevStatusch <- message

            case ACK:
                // ACKNOWLEDGE

            case SendMatrix:
                MatrixUpdatech <- message


            case UpdatedMatrix:
                MatrixUpdatech <- message
            
            message.Done = true
          }
        }
      }


        /*
      // Check if message has been processed.
      if !msgRecieved.Done {
        message := msgRecieved

        // If select = 1, new order was recieved.
        if message.Select == 1 {
          // Sett info inn på message
        }

      // Wait to everyone agree

      // Send message to local ordermanager*/
      fmt.Println("Message Recievied")




      //}

    // --------------------------------------------------------------------------Case triggered by update in peers
    case p := <- syncChans.PeerUpdate:

    fmt.Println("New peer: ", p.New)
    if len(p.New) > 0 {
        newID, _ := strconv.Atoi(p.New) // ID of new Peer
        
        if newID == elevatorConfig.ElevID && len(p.Peers) == 1 {
            // You are alone on network (Either first or someone disappeard)
            // do nothing
            elevatorConfig.IsDefined = true
            Online = true
            fmt.Println("HEYO, I AM", newID, "AND IM FIRST")
        } else if newID == elevatorConfig.ElevID && len(p.Peers) > 1 {
            // Either been offline or first time online
            // Ask for matrix
            elevatorConfig.IsDefined = true
            Online = true
            fmt.Println("HEYO, I AM ", newID, "AND IM ONLINE")

        } else if newID != elevatorConfig.ElevID && elevatorConfig.IsDefined{
            // Already online, send matrix to new

            message := Message{Select: SendMatrix, ID: newID}
            MatrixUpdatech <- message
        }
    }


    for _, peerLost := range p.Lost {
        newID, _ := strconv.Atoi(peerLost)
        if newID != elevatorConfig.ElevID {
            // Someone else is offline
            message := Message{Select: UpdateOffline, ID: newID}
            UpdateElevStatusch <- message
        } else {
          Online = false
          fmt.Println("I AM OFFLINE! BUT CHILL, I GOT THIS")
        }
    }

    }
  }
}
