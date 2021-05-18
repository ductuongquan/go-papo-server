# Developer guide for Papo Server

This section cover index of developer guide for Papo Server

## import rulers

* Alway using these description lines for internal dependencies import

```
import (
	/**
	 * External dependencies
	 */
	'os'

	/**
	 * Internal dependencies
	 */
	'bitbucket.org/enesyteam/papo/cmd/commands'
)
```

### import path

You need to use the **complete import path** `github.com/levin/foo` if you want `go get` to work, and you should do it this way if you expect other people to use your package.

## Coding styles

* Alway use *space* between any bracket:

Bad:
```
func abc(x){
	...
}
```

Good:
```
func abc( x ) {
	...
}
```


