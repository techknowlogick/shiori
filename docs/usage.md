Before using `shiori`, make sure it has been installed on your system. By default, `shiori` will store its data in directory `$HOME/.local/share/shiori`. If you want to set the data directory to another location, you can set the environment variable `SHIORI_DIR` to your desired path.

- [Running Docker Container](#running-docker-container)
- [Using Command Line Interface](#using-command-line-interface)
- [Using Web Application](#using-web-application)
- [CLI Examples](#cli-examples)

## Running Docker Container

> If you are not using `shiori` from Docker image, you can skip this section.

After building the image you will be able to start a container from it. To
preserve the data, you need to bind the directory for storing database and thumbnails. In this example we're binding the data directory to our current working directory :

```
docker run -d --rm --name shiori -p 8080:8080 -v $(pwd):/srv/shiori techknowlogick/shiori
```

The above command will :

- Creates a new container from image `techknowlogick/shiori`.
- Set the container name to `shiori` (option `--name`).
- Bind the host current working directory to `/srv/shiori` inside container (option `-v`).
- Expose port `8080` in container to port `8080` in host machine (option `-p`).
- Run the container in background (option `-d`).
- Automatically remove the container when it stopped (option `--rm`).

After you've run the container in background, you can access console of the container :

```
docker exec -it shiori sh
```

Now you can use `shiori` like normal. If you've finished, you can stop and remove the container by running :

```
docker stop shiori
```

## Using Command Line Interface

```
Simple command-line bookmark manager built with Go

Usage:
  shiori [command]

Available Commands:
  account     Manage account for accessing web interface
  add         Bookmark the specified URL
  delete      Delete the saved bookmarks
  export      Export bookmarks into HTML file in Netscape Bookmark format
  help        Help about any command
  import      Import bookmarks from HTML file in Netscape Bookmark format
  open        Open the saved bookmarks
  pocket      Import bookmarks from Pocket's exported HTML file
  print       Print the saved bookmarks
  search      Search bookmarks by submitted keyword
  serve       Serve web app for managing bookmarks
  update      Update the saved bookmarks

Flags:
  -h, --help   help for shiori

Use "shiori [command] --help" for more information about a command.
```

## Using Web Application

To access web application, you need to have at least one account. To create new account, run this command :

```
shiori account add <your-desired-username>
Password: <enter-your-password>
```

If you are using Docker container, you can access the web application immediately in `http://localhost:8080`. If not, you need to run `shiori serve` first.

## CLI Examples

1. Save new bookmark with tags "nature" and "climate change".

   ```
   shiori add https://grist.org/article/let-it-go-the-arctic-will-never-be-frozen-again/ -t nature,"climate change"
   ```

2. Print all saved bookmarks.

   ```
   shiori print
   ```

2. Print bookmarks with index 1 and 2.

   ```
   shiori print 1 2
   ```

3. Search bookmarks that contains "sqlite" in their title, excerpt, url or content.

   ```
   shiori search sqlite
   ```

4. Search bookmarks with tag "nature".

   ```
   shiori search -t nature
   ```

5. Delete all bookmarks.

   ```
   shiori delete
   ```

6. Delete all bookmarks with tag "nature".

   ```
   shiori delete $(shiori search -t nature -i)
   ```

7. Update all bookmarks' data and content.

   ```
   shiori update
   ```

8. Update bookmark in index 1.

   ```
   shiori update 1
   ```

9. Change title and excerpt from bookmark in index 1.

   ```
   shiori update 1 -i "New Title" -e "New excerpt"
   ```

10. Add tag "future" and remove tag "climate change" from bookmark in index 1.

    ```
    shiori update 1 -t future,"-climate change"
    ```

11. Import bookmarks from HTML Netscape Bookmark file.

    ```
    shiori import exported-from-firefox.html
    ```

12. Export saved bookmarks to HTML Netscape Bookmark file.

    ```
    shiori export target.html
    ```

13. Open all saved bookmarks in browser.

    ```
    shiori open
    ```

14. Open text cache of bookmark in index 1.

    ```
    shiori open 1 -c
    ```

15. Serve web app in port 9000.

    ```
    shiori serve -p 9000
    ```

16. Create new account for login to web app.

    ```
    shiori account add username
    ```
