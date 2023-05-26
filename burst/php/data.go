package php

import "path"

var PhpPath = path.Join("/usr", "local", "bin", "php")

var MockPhp = "<html><body><?php echo \"Hello World!\" ?></body></html>"
var MockPhpResult = "<html><body>Hello World!</body></html>"
