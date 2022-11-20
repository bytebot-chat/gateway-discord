#! /usr/bin/env python
import argparse
import logging
import redis
import uuid
import json


def respondToChannelOrThread(message, source, content, should_reply=False, should_mention=False):
    # debug print keys in message
    # for key in message:
    #    logging.info("Key: %s, Value: %s\n", key, message[key])

    msg = dict({})
    metadata = dict({})
    metadata['source'] = source
    metadata['dest'] = 'discord'
    metadata['id'] = str(uuid.uuid4())
    msg['metadata'] = metadata

    msg['channel_id'] = message['message']['channel_id']
    msg['content'] = content
    msg['previous_message'] = message['message']
    msg['should_reply'] = should_reply
    msg['should_mention'] = should_mention

    return msg


def main():
    # Parse command line arguments
    parser = argparse.ArgumentParser(description='Ping Pong')
    parser.add_argument("-r", "--redis", type=str,
                        default="localhost", help="Redis server address")
    parser.add_argument("-p", "--port", type=int,
                        default=6379, help="Redis server port")
    parser.add_argument('-i', '--inbound', type=str,
                        default='discord-inbound', help='Inbound queue name')
    parser.add_argument('-o', '--outbound', type=str,
                        default='discord-outbound', help='Outbound queue name')

    args = parser.parse_args()

    # Configure logger
    logging.basicConfig(level=logging.INFO)
    logging.info("Starting ping pong")

    # Connect to Redis
    r = redis.Redis(host=args.redis, port=6379, db=0)
    p = r.pubsub()
    p.subscribe(args.inbound)

    # Main loop
    while True:
        message = p.get_message()
        if message and message['type'] == 'message':
            # logging.info("Received message: %s", message['data'])

            # Parse message
            msg = json.loads(message['data'])

            # for key in msg:
            #    logging.info("Key: %s, Value: %s\n", key, msg[key])

            # Check if it's a ping
            if msg['message']['content'] == 'ping':
                # Respond with a pong
                logging.info("Responding to ping")
                pong = respondToChannelOrThread(
                    msg, 'python-pingpong', 'pong from python', should_reply=True)
                r.publish(args.outbound, json.dumps(pong))


if __name__ == "__main__":
    main()
