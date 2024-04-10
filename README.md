# Cosign key generation - Compiled to WASM

This is a tiny go project inteded for use in the web browser WASM that does key generation for signing artifacts with [sigstore/cosign](github.com/sigstore/cosign). This project consists of code _respectfully stolen_ from the [containers/image](https://github.com/containers/image/) project, and some glue code for tying it in with WASM and JS. 

The WASM built from this project is used in production in the [BlueBuild Workshop](https://github.com/blue-build/workshop).

## Building locally

```sh
devbox shell # optional, you can also manually install tinygo
tinygo build -o cosign.wasm -target wasm main.go
```

## Usage

1. Get the supplies, two options:
    - Manually
        - Build the project into WASM using the command above.
        - Get the `wasm_exec.js` file using the following command `cp "$(tinygo env GOROOT)/misc/wasm/wasm_exec.js" ./`.
    - From GitHub releases
        - Download the `cosign.wasm` and `wasm_exec.js` files from the latest GitHub release.
2. Copy these files into the folder for static files in your web development project.
3. Add and adapt the following code into your project:
    - Add `<script src="/wasm_exec.js" defer></script>` in the `<head>`
    - In another script, copy the following boilerplate:
        ```js
        const go = new Go();

        // You can prefetch the WASM file, if you want to
        WebAssembly.instantiateStreaming(fetch("/cosign.wasm"), go.importObject).then(
            async (obj) => {
                const wasm = obj.instance;
                go.run(wasm);
                // The Go code sets these global variables
                // Make sure to empty them after use to prevent key leakage
                console.log(cosignPublicKey);
                console.log(cosignPrivateKey);
            }
        );
        ```