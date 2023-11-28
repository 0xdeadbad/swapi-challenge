package api

type Person struct {
	Pid    int    `json:"id,omitempty" redis:"id"     bson:"id"`
	Height string `json:"height"       redis:"height" bson:"height"`
	Name   string `json:"name"         redis:"name"   bson:"name"`
	Gender string `json:"gender"       redis:"gender" bson:"gender"`
}
