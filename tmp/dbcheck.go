package main
import (
  "database/sql"
  "fmt"
  _ "modernc.org/sqlite"
)
func main(){
  db, err := sql.Open("sqlite", "gitforge.db")
  if err != nil { panic(err) }
  defer db.Close()

  var users int
  if err := db.QueryRow("select count(*) from users").Scan(&users); err != nil { panic(err) }
  var repos int
  if err := db.QueryRow("select count(*) from repositories").Scan(&repos); err != nil { panic(err) }
  fmt.Printf("users=%d repos=%d\n", users, repos)

  rows, err := db.Query("select id, username, email from users order by id desc limit 10")
  if err != nil { panic(err) }
  defer rows.Close()
  fmt.Println("-- users --")
  for rows.Next(){
    var id int64; var u,e string
    _ = rows.Scan(&id,&u,&e)
    fmt.Printf("%d\t%s\t%s\n", id,u,e)
  }

  rows2, err := db.Query("select r.id, u.username, r.name, r.visibility from repositories r join users u on u.id=r.owner_id order by r.id desc limit 10")
  if err != nil { panic(err) }
  defer rows2.Close()
  fmt.Println("-- repos --")
  for rows2.Next(){
    var id int64; var o,n,v string
    _ = rows2.Scan(&id,&o,&n,&v)
    fmt.Printf("%d\t%s/%s\t%s\n", id,o,n,v)
  }
}
