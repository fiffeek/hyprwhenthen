#!/bin/bash

ADDRESS="0x$1"

hyprctl dispatch togglefloating "address:$ADDRESS" &&
  hyprctl dispatch resizewindowpixel exact 50% 50%,"address:$ADDRESS" &&
  hyprctl dispatch centerwindow
