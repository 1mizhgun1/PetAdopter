// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Misha",
            "url": "http://t.me/KpyTou_HocoK_tg"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/user/login": {
            "post": {
                "description": "login",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Login",
                "operationId": "login",
                "parameters": [
                    {
                        "description": "request",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "response 200",
                        "schema": {
                            "$ref": "#/definitions/handlers.LoginResponse"
                        }
                    },
                    "400": {
                        "description": "response 400\" \"invalid",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "response 500\" \"internal",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/logout": {
            "post": {
                "description": "logout",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Logout",
                "operationId": "logout",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "401": {
                        "description": "Unauthorized"
                    }
                }
            }
        },
        "/user/signup": {
            "post": {
                "description": "Add a new user to the database",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Sign up",
                "operationId": "sign-up",
                "parameters": [
                    {
                        "description": "request",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.SignUpRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "response 200",
                        "schema": {
                            "$ref": "#/definitions/handlers.SignUpResponse"
                        }
                    },
                    "400": {
                        "description": "response 400\" \"invalid",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "response 500\" \"internal",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handlers.LoginRequest": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handlers.LoginResponse": {
            "type": "object",
            "properties": {
                "refresh_token": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/user.User"
                }
            }
        },
        "handlers.SignUpRequest": {
            "type": "object",
            "properties": {
                "locality_id": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "handlers.SignUpResponse": {
            "type": "object",
            "properties": {
                "refresh_token": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/user.User"
                }
            }
        },
        "user.User": {
            "type": "object",
            "properties": {
                "locality_id": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "PetAdopter API",
	Description:      "API server for PetAdopter.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
