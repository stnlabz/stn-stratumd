# dw-stratumd

# Setting Up
- go init stratumd
- go mod tidy

# Starting up
- go build
- ./stratumd

# Serviving system reboots
- # add @reboot job (uses your $HOME path)
( crontab -l 2>/dev/null; \
  echo "@reboot /bin/bash -lc 'cd $HOME/stratumd && nohup ./stratumd >> $HOME/stratumd/logs/stratumd.log 2>&1 &'" \
) | crontab -