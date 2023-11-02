import datetime
import pathlib
import time


def main():
    while True:
        print("Python program is working")
        time.sleep(1)
        pathlib.Path("test.txt").write_text(
            datetime.datetime.now(tz=datetime.timezone.utc).isoformat()
        )


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("Python program interrupted")
