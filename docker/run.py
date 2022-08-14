#!/usr/bin/env python3

import os
import sys
sys.path.insert(0, "/code")
from app import app

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8080))
    app.wisdom = "Hello from the (much better) Nyan Cat!"
    if len(sys.argv) > 1:
        app.wisdom = " ".join(sys.argv[1:])
    app.run(host='0.0.0.0', port=port)
