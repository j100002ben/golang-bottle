package bottle

import (
    "container/list"
    "time"
)

type Object interface {}

type BottleRecycler struct {
    get chan Object
    give chan Object
}

type Bottle struct {
    created time.Time
    object Object
}

func MakeBottleRecycler(maker func() Object, waitTime time.Duration) (bc BottleRecycler) {
    bc = BottleRecycler{get: make(chan Object), give: make(chan Object)}
    go func() {
        storeList := new(list.List)
        for {
            if storeList.Len() == 0 {
                storeList.PushFront(Bottle{created: time.Now(), object: maker()})
            }

            element := storeList.Front()

            timeout := time.NewTimer(waitTime)

            select {
            case object := <- bc.give:
                timeout.Stop()
                storeList.PushFront(Bottle{created: time.Now(), object: object})

            case bc.get <- element.Value.(Bottle).object:
                timeout.Stop()
                storeList.Remove(element)

            case <- timeout.C:
                element := storeList.Front()
                for element != nil {
                    next := element.Next()
                    if time.Since(element.Value.(Bottle).created) > waitTime {
                        storeList.Remove(element)
                        element.Value = nil
                    }
                    element = next
                }
            }
        }
    }()

    return
}

func (bc BottleRecycler) Get() (object Object) {
    return <- bc.get
}

func (bc BottleRecycler) Give(object Object) {
    bc.give <- object
}