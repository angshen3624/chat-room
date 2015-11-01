package main

import "fmt"
import "os"
import "strconv"
import "net"
import "io"
import "strings"

type ClientStruct struct {
    ID int
    MODE int
    msg []byte
}



func main() {

    arg_num := len(os.Args)
    fmt.Printf("number of arguments: %d\n", arg_num)//
    if arg_num != 2 {
        fmt.Println("number of arugments should be exactly 4")
        os.Exit(0)
    }
    port_str_temp := []string{":", os.Args[1]}
    port_str := strings.Join(port_str_temp, "")
    port_int, err := strconv.Atoi(os.Args[1]) 
    if err != nil{
        fmt.Println("error happened ,exit\n")
        os.Exit(0)
    }
    fmt.Printf("Port is: %d\n", port_int)
    
    tcp, err := net.Listen("tcp", port_str)
    if err != nil{
        fmt.Printf("cannot listen at port %d\n", port_int)
        os.Exit(0)
    }

    fmt.Println("Starting the server\n")//
    var clientCount int = -1
    chnl := make(chan ClientStruct)  
  
    chnl_conn:=make(chan net.Conn)
    chnl_ID:=make(chan int)
    chnl_delete:=make(chan int)

    go msgManager(chnl, chnl_conn, chnl_ID, chnl_delete)
    for {
        conn, err := tcp.Accept()
        if err != nil{
            fmt.Printf("Cannot accept at port %d\n", port_int)
            os.Exit(0)
        }
        clientCount += 1
        fmt.Printf("Creating client ID: %d\n", clientCount)//
        chnl_conn<-conn
        chnl_ID<-clientCount
        fmt.Println("Accepted the connection")//
        go eachConnection(conn, clientCount, chnl, chnl_delete)
    }


}

func msgManager(chnl chan ClientStruct, chnl_conn chan net.Conn, chnl_ID chan int, chnl_delete chan int){
    mp:=make(map[int]net.Conn)
    for{
        select{
            case clientID_delete := <- chnl_delete:
                delete(mp, clientID_delete)

            case conn := <- chnl_conn:
                select{
                    case clientID:=<- chnl_ID:
                        mp[clientID]=conn

                }

            case myClient := <- chnl :
        //        msg, msgMode := msgFormer(myClient)
                if myClient.MODE == -1{
                    for i := 0; i < len(mp); i++ {  // FIXME choose from valid connections
                        mp[i].Write(myClient.msg) 
                    }
                }else {
                mp[myClient.MODE].Write(myClient.msg) 
                }
        }
    }
}


func msgFormer(myClient ClientStruct) (ClientStruct) {
    SemicExist := 0
    AllExist := 0
    SemicIndex := -1
    Whoami := 0
    ID_str := strconv.Itoa(myClient.ID) 
    var msgMode int = -1
    msg := myClient.msg

    for i:= range []byte(msg){
        if msg[i] == 58 {
            SemicExist = 1
            SemicIndex = i
        } else if msg[i] == 97 && msg[i+1] == 108 && msg[i+2] == 108{
            AllExist = 1
        } else if msg[i] == 119 && msg[i+1] == 104 && msg[i+2] == 111 && msg[i+3] == 97 && msg[i+4] == 109 && msg[i+5] == 105{
            Whoami = 1
        }
    }
    var s []string
    if Whoami == 1{
        msgMode = myClient.ID
        s = []string{"chitter:", ID_str, "\n"}
    }else if SemicExist == 0 {
        msgMode = -1
        s = []string{ID_str, ": ", string(msg)}
    }else if (SemicExist & AllExist) == 1 {
        msgMode = -1
        s = []string{ID_str, ": ", string(msg[(SemicIndex+1):])}
    }else if (SemicExist == 1 && AllExist == 0){
        msgMode, _ = strconv.Atoi(string(msg[0:SemicIndex]))
        s = []string{ID_str, ": ", string(msg[(SemicIndex+1):])}
    }else{
        fmt.Println("Message string not supported")
    }
    v := ClientStruct{myClient.ID, msgMode,[]byte(strings.Join(s, ""))}

    return v
}

func eachConnection(conn net.Conn, clientID int, chnl chan ClientStruct, chnl_delete chan int){
    buf := make([]byte, 1000)
    defer conn.Close()
    //defer 
    //defer delete(mp, clientID)

    for {
        n, err := conn.Read(buf);
         //        msg, msgMode := msgFormer(myClient)
        if err == io.EOF{
            fmt.Printf("Client %d disconnected\n", clientID);
            chnl_delete <- clientID
            return
        } else if err != nil{
            fmt.Printf("Error: Reading data : %s \n", err);
            return
        } else{
            var v ClientStruct
            v = ClientStruct{clientID, 0, buf[0:n]}
            vv := msgFormer(v)
            chnl <- vv
        }
    }
}