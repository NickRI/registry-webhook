package webhook

import "errors"

//ErrBadChunksSize used to present chunks count
var ErrBadChunksSize = errors.New("Error bad chunk size presented")
