[![Go Report](https://goreportcard.com/badge/github.com/masonkmeyer/agify)](https://goreportcard.com/badge/github.com/masonkmeyer/agify)
![Build](https://github.com/github/docs/actions/workflows/build.yml/badge.svg)

# Agify 
 
 Agify is a go client for the [agify.io](https://agify.io/) api. 

 ## Examples

 You can use this library to call the API client. 
 
 ```golang
client := agify.NewClient()
prediction, rateLimit, err := client.Predict("michael")
 ```

This client also supports predictions by country and batch predictions. 
