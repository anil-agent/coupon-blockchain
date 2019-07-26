# coupon-blockchain
Blockchain solution for Coupon Management

Commands for interacting with the chaincode

Installing the chaincode
docker exec cli peer chaincode install -n chaincodename -p PathtoChaincodefile -v v0

Instantiating the chaincode on the Channel

docker exec cli peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n chaincodename -c '{"function":"initLedger","Args":[""]}'

Invoking the chaincode

1. Create Coupon

 docker exec cli peer chaincode invoke -C channelname -n chaincodename -c '{"Args":["createCoupon", "Big Sale", "10-07-2019", "31-12-2019", "10.5", "Ankit1","Ankit1@gmail.com"]}'

2. Get Coupon by CouponID

docker exec cli peer chaincode invoke -C channelname -n chaincodename -c '{"Args":["getCouponById","couponID"]}'
