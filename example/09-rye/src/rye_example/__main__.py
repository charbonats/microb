import time


if __name__ == "__main__":
    while True:
        print("Hello, World!")
        try:
            time.sleep(1)
        except KeyboardInterrupt:
            break
