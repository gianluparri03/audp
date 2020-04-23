# AUDP

**A**n **U**nnamed **D**omotic **P**roject.


## Summary

- [How does it works](#How-does-it-works)
- [How to run](#How-to-run)
  - [Recap](#Recap)
  - [API](#API)
- [Authors](#Authors)
- [License](#License)


## How does it works

### Recap

So, how does it work.
Let's start from the beginning. A controller, when plugged in, makes a post request to `server/controllers/wakeup`, with a body content like `{"mac_address": "20:1A:06:98:7B:64"}` (side note: like in the following post requests, `Content-Type` must be set to `application/json`).
*PS: Actually, it doesn't matter if it really is a mac address: the `mac_address` field is just a field used to uniquely identify that device. For example, an arduino registers into the server with the field set to `NoobMaster69`; when I shut down the arduino, once restarted it won't remember anything, unless `NoobMaster69`. Then it looks in the server if there is a controller with `mac_address` set to `NoobMaster69`already registered: if there is, it means it was registered, once, so it has some configs saved in the server and it can recover them. Otherwise it has to register.
So, in the end, you can set `mac_address` to whatever you want. I recommend to set it to your real mac address.*
By the way, we were saying: it makes this post request. If there is a sleeping controller in the server with that `mac_address`, it wakes it up. Otherwise, the controller must register at `server/controllers/add`: the necessary fields are `{"name": "Name", "mac_address": "20:1A:06:98:7B:64", "port": 80}`; you can also create a new device immediatly, by passing it as it follows `{"devices": [{"name": "Light", "status": 0, "GPIO": 3}]}`.
*NB: for controllers, `name` and `mac_address` are unique, and you can only register a controller per ip. Device's `name`s are unique too. You can't set more devices to work in the same controller in the same GPIO.*

You may ask: what is GPIO? Is it really what I'm thinking?
Yes, the server saves in which gpio the device is connected. Again, you may ask: why? Well, we said that all the data will be saved in the server. So, all the devices managed by a controller are saved in the server so that the server, if gone to sleep, can retrieve back all his configs.
In fact, if a device wakes up, the server's response will contain all the old configs. And, if there was a device turned on, it will immediatly turn it on. Nice, isn't it?

Ok, but, when a controller falls alseep?
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
