# AUDP

**A**n **U**nnamed **D**omotic **P**roject.


## Summary

- [How does it works](#How-does-it-works)
  - [Controllers](#Controllers)
- [How to run](#How-to-run)
  - [API](#API)
- [Authors](#Authors)
- [License](#License)


## How does it works

### Controllers

Let's start from the beginning. A controller, when powered on, makes a post request to `/controllers/wakeup`, with a body content like `{"code": "my_unique_code"}`, with `Content-Type` set to `application/json`.

> What's `code`?
It's just a field used to uniquely identify that device. I recommend you to use your controller's MAC address as the code, but you can set it to whatever you want.

If there is a sleeping controller in the server with that `code`, it wakes it up. Otherwise, the controller must register at `/controllers/add`: the necessary fields are `"name"` (string), `"code"` (string) and `port` (int).
*NB: for controllers, `name` and `code` are unique, and you can only register a controller per ip*

> Ok, but, when a controller falls alseep?
Every 10 seconds, the server pings all the controllers at the given port. If it can't connect to that ip, it sets the controller's `sleeping` value to `true`.

## How to run

### API

```bash
# Clone the repo
git clone https://github.com/gianluparri03/audp.git
cd audp/api

# Start the api
go run *.go
```


## Authors

- **Parri Gianluca** - [@gianluparri03](https://github.com/gianluparri03)

Click on [this link](https://github.com/gianluparri03/audp/graphs/contributors) to see the list of contributors who participated in this project.


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for more details.
