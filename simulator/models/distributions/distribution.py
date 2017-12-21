from .exponential import Exponential 
from .normal import Normal

class Distribution(object):
    
    def __init__(self, conf, distribution_name, key_prefix):
        self.distribution = None
        
        if(distribution_name == 'exponential'):
            avg = float(conf[key_prefix + '_avg'])
            self.distribution = Exponential(avg)
            
        elif(distribution_name == 'normal'):
            avg = float(conf[key_prefix + '_avg'])
            desviation = float(conf[key_prefix + '_desv'])
            self.distribution = Normal(avg, desviation)
    
    def get_next_value(self):
        return self.distribution.get_next_value()
