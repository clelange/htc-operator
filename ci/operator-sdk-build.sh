#!/bin/bash
ls
curl -sSL 'https://github.com/operator-framework/operator-sdk/releases/download/v0.17.0/operator-sdk-v0.17.0-x86_64-linux-gnu' > operator-sdk
chmod +x operator-sdk
./operator-sdk build randname
