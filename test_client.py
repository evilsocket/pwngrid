#!/usr/bin/env python3
import base64
import hashlib
import requests
from Crypto.Signature import PKCS1_PSS
from Crypto.PublicKey import RSA
import Crypto.Hash.SHA256 as SHA256

# ssh-keygen -t rsa -m PEM -b 4096 -f test-unit
priv_path = 'test-unit'
pub_path = priv_path + '.pub'

name = 'test-unit'

with open(pub_path) as fp:
    pubkey = RSA.importKey(fp.read())

with open(priv_path) as fp:
    privkey = RSA.importKey(fp.read())

pub_key_pem = pubkey.exportKey('PEM').decode("ascii")

# python is special
if 'RSA PUBLIC KEY' not in pub_key_pem:
    pub_key_pem = pub_key_pem.replace('PUBLIC KEY', 'RSA PUBLIC KEY')
# print("'%s'" % pub_pem)
fingerprint = hashlib.sha256(pub_key_pem.encode("ascii")).hexdigest()

# print("PEM")
# print(pub_key_pem)

pub_key_pem_b64 = base64.b64encode(pub_key_pem.encode("ascii")).decode("ascii")

identity = '%s@%s' % (name, fingerprint)

hasher = SHA256.new(identity.encode("ascii"))
signer = PKCS1_PSS.new(privkey, saltLen=16)
signature = signer.sign(hasher)
signature_b64 = base64.b64encode(signature).decode("ascii")

# print("hash(data) = %s" % ''.join(["%02x" % b for b in hasher.digest()]))
# print("signature  = %s" % ''.join(["%02x" % b for b in signature]))
api_address = 'http://api.pwnagotchi.ai/api/v1/unit/enroll'
enroll = {
    'identity': identity,
    'public_key': pub_key_pem_b64,
    'signature': signature_b64
}

print("enrolling to %s :\n\n%s\n" % (api_address, enroll))
r = requests.post(api_address, json=enroll)

print("%d" % r.status_code)
print(r.json())
