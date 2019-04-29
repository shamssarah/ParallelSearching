package main // templates

import (
        "fmt"
        "html/template"
        "io/ioutil"
        "net/http"
        // "regexp"
        "bufio"
        // "fmt"
        "io"
        "log"
        "net"
        // "os"
        "strings"
         // "math/rand"
)



//Client represents a client
type Client struct {
        conn     net.Conn
        nickname string
        ch       chan string
        files    [] string
}


//ReadLinesInto is a method on Client type
//it keeps waiting for user to input a line, ch chan is the msgchannel
//it formats and writes the message to the channel
func (c Client) ReadLinesInto(ch chan<- string) {
        bufc := bufio.NewReader(c.conn)
        for {
                line, err := bufc.ReadString('\n')
                if err != nil {
                        break
                }
                ch <- fmt.Sprintf("%s: %s", c.nickname, line)
        }
}

//WriteLinesFrom is a method
//each client routine is writing to channel
func (c Client) WriteLinesFrom(ch <-chan string) {
        for msg := range ch {
                _, err := io.WriteString(c.conn, msg)
                if err != nil {
                        return
                }
        }
}



func (c Client) AssigningFileFromChunk(conn net.Conn, visited [] string) {

        // ind := rand.Intn(len(c.files) - 0) + len(c.files)
         fmt.Println("c files %s%s", c.files[0], c.files[0])
        new_file_to_be_processed := c.files[0]
        visited = append(visited,new_file_to_be_processed)
        fmt.Println("please work%s", new_file_to_be_processed)
        io.WriteString(conn,new_file_to_be_processed) 
        //ch <- new_file_to_be_processed

        // for msg := range ch {
        //         // _, err := io.WriteString(c.conn, msg)
        //         // if err != nil {
        //                 return 
        //         }
        // }
}


func promptNick(c net.Conn, bufc *bufio.Reader) string {
        io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
        io.WriteString(c, "What is your nick? ")
        nick, _, _ := bufc.ReadLine()
        return string(nick)
}



func handleMessages(msgchan <-chan string, addchan <-chan Client, rmchan <-chan Client) {
        clients := make(map[net.Conn]chan<- string)

        for {
                select {
                case msg := <-msgchan:
                        log.Printf("\nNew message: %s", msg)


                        // log.Printf ("IIIIIIIIIII %s",msg)
                        // for _, ch := range clients {
                        //         go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m" }(ch)
                        // }
                case client := <-addchan:
                        log.Printf("New client: %v\n", client.conn)
                        clients[client.conn] = client.ch
                case client := <-rmchan:
                        log.Printf("Client disconnects: %v\n", client.conn)
                        delete(clients, client.conn)
                }
        }
}

//Page represents a Page
type Page struct {
        Title string
        Body  []byte
}

func (pageToSave Page) savePage() error {
        fileName := pageToSave.Title + ".txt"
        err := ioutil.WriteFile(fileName, pageToSave.Body, 0700)
        return err
}

func loadPage(pageToLoad string) (*Page, error) {
        fileName := pageToLoad + ".txt"
        fileBody, err := ioutil.ReadFile(fileName)
        if err != nil {
                return nil, err
        }
        return &Page{Title: pageToLoad, Body: fileBody}, err
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hi there, This is defult page%s!", r.URL.Path[1:])
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
        pageTitle := r.URL.Path[len("/view/"):]
        loadedPage, err := loadPage(pageTitle)
        if err != nil {
                http.Redirect(w, r, "/edit/"+pageTitle, http.StatusFound)

        } else {
                viewTemplate, _ := template.ParseFiles("view.html")
                viewTemplate.Execute(w, loadedPage)
                //fmt.Fprintf(w, "<html><body>%s</body></html> ", loadedPage.Body)
        }
}

func editHandler(w http.ResponseWriter, r *http.Request) {
        pageTitle := r.URL.Path[len("/edit/"):]
        loadedPage, err := loadPage(pageTitle)
        if err != nil {
                loadedPage = &Page{Title: pageTitle}
        }

        editTemplate, _ := template.ParseFiles("edit.html")
        editTemplate.Execute(w, loadedPage)
        /*
                fmt.Fprintf(w, "<h1>Editing %s</h1>"+
                        "<form action=\"/save/%s\" method=\"POST\">"+
                        "<textarea name=\"body\">%s</textarea><br>"+
                        "<input type=\"submit\" value=\"Save\">"+
                        "</form>",
                        loadedPage.Title, loadedPage.Title, loadedPage.Body)
        */
}

func promptFilenames (c net.Conn, bufc *bufio.Reader) []string {

        filenames, _, _ := bufc.ReadLine()
        output := make([]string,1)
        output[0] = string(filenames)

        log.Println("\nFILENAMES:\n")
        for  i := 1; i < 20; i++ {
                filenames, _, _ := bufc.ReadLine()
                // log.Println(string(filenames),"\n")
                output = append (output,string(filenames))
        }
        log.Println(output[3])
        return output

}


func handleConnection(c net.Conn, msgchan chan<- string, addchan chan<- Client, rmchan chan<- Client , qchan chan <- string ,visited [] string) {
        bufc := bufio.NewReader(c)
        defer c.Close()

        //we first need to add current client to the channel
        //filling in the client structure
        client := Client{
                conn:     c,
                nickname: promptNick(c, bufc),
                ch:       make(chan string),
                files:    promptFilenames(c,bufc),
        }


        // log.Printf ("filenamesssssssssssssssssssssssssssssssss\n",client.files[0],"\n")
        if strings.TrimSpace(client.nickname) == "" {
                io.WriteString(c, "Invalid Username\n")
                return
        }

        // Register user, our messageHandler is waiting on this channel
        // it populates the map
        addchan <- client

        //ignore for the time being
        defer func() {
                msgchan <- fmt.Sprintf("User %s left the chat room.\n", client.nickname)
                log.Printf("Connection from %v closed.\n", c.RemoteAddr())
                rmchan <- client
        }()

        //just a welcome message
        // io.WriteString(c, fmt.Sprintf("Wllllllllllle, %s!\n\n", client.nickname))

        //We are now populating the other channel now
        //our message handler is waiting on this channel as well
        //it reads this message and copies to the individual channel of each Client in map
        // effectively the broadcast
        msgchan <- fmt.Sprintf("New user %s has joined the chat room.\n", client.nickname)
        


        client.AssigningFileFromChunk(c, visited) //channel written



        
        // another go routine whose purpose is to keep on waiting for user input
        //and write it with nick to the
        go client.ReadLinesInto(msgchan)

        //given a channel, writelines prints lines from it
        //we are giving here client.ch and this routine is for each client
        //so effectively each client is printitng its channel
        //to which our messagehandler has added messages for boroadcast
        client.WriteLinesFrom(client.ch)
}




func acceptLoop(l net.Listener,passwordchan chan string ) {
        defer l.Close()
        fmt.Println ("In Accept LOop: ")
        pass := <- passwordchan
        fmt.Println ("PASSSSSSSSWORDDDDD!!!!!!!\t",pass)

        visited := make([] string ,1)

        msgchan := make(chan string)
        procchan := make (chan string) // to be processed filename

        //A channel to keep track of Clients, clients are added to this channel
        //handleMessages then iterates through clients and for each appends to its channel
        //the messages sent by other users, broadcat
        addchan := make(chan Client)
        rmchan := make(chan Client)

        //this function has only one instance and is a goroutine
        go handleMessages(msgchan, addchan, rmchan)

        for {
                conn, err := l.Accept()
                if err != nil {
                        fmt.Println(err)
                        continue
                }
                //for each client, we do have a separate handleConnection goroutine
                go handleConnection(conn, msgchan, addchan, rmchan, procchan ,visited)
        }

}
 

func Save(passwordchan chan string) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // logger <- r.Host
        // io.WriteString(w, "Hello world!")

        pageTitle := r.URL.Path[len("/save/"):]
        pageBody := r.FormValue("body")

        passwordchan <- fmt.Sprintf("Password %s\n",pageBody)

        var pageToSave = Page{Title: pageTitle, Body: []byte(pageBody)}
        pageToSave.savePage()
    }
}

func main() {

       
    passwordchan := make(chan string)

    saveHandler := Save(passwordchan)
    
     
    http.HandleFunc("/", defaultHandler)
    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler) 
    http.HandleFunc("/save/", saveHandler)
    // http.HandleFunc("/save/", MustParams(saveHandler,passwordchan))


    listener, err := net.Listen("tcp", ":3000")
    if err != nil {
        log.Fatal(err) 
        fmt.Println ("Error: ",err)
    }
    // listener2, err := net.Listen("tcp", "localhost:8081")
    // if err != nil {
    //     log.Fatal(err)
    // }

    go acceptLoop(listener,passwordchan) // slaves wait
    http.ListenAndServe(":8081", nil)
    // acceptLoop(listener2)  // run in the main goroutine wait for webserver
    
}




