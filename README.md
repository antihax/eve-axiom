<span class="badge-buymeacoffee">
<a href="https://www.buymeacoffee.com/antihax" title="Donate to this project using Buy Me A Coffee"><img src="https://img.shields.io/badge/buy%20me%20a%20coffee-donate-yellow.svg" alt="Buy Me A Coffee donate button" /></a>
</span>

# eve-axiom

Provides a killmail to attribute service to resolve dogma into a json output of a fittings capability.

## dockerized

`docker run antihax/eve-axiom -p 3005:3005`

## compilation

* Compile [libdogma](https://github.com/antihax/libdogma) and place it in the necessary paths for cgo to find.
* Build the cmd directory.

## operation

POST ESI formatted JSON killmail at the `:3005/killmail` endpoint and receive a JSON response.
The `:3000` port has prometheus stats and golang pprof information. This port should not be exposed, please protect it.

