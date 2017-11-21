import configparser, logging

def getConfig(path_file, conf_section):
    config_parser = configparser.ConfigParser()
    config_parser.read(path_file)
    required_conf = config_parser[conf_section]
    return required_conf

def getLogger(path_file, logger_name):
    handler = logging.FileHandler(path_file)
    logger = logging.getLogger(logger_name)
    logger.setLevel(logging.INFO)
    logger.addHandler(handler)
    return logger