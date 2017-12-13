import abc

class Distribution(object):
    __metaclass__ = abc.ABCMeta
    
    @abc.abstractmethod
    def get_next_value(self):
        return
