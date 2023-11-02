import datetime
import pathlib
import time


def main():
    while True:
        print("Python app 2 is still running")
        time.sleep(1)
        pathlib.Path("test-app-1.txt").write_text(
            datetime.datetime.now(tz=datetime.timezone.utc).isoformat()
        )


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("Python app 2 interrupted")
