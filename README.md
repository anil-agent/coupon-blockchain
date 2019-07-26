# coupon-blockchain
Blockchain solution for Coupon Management


Create Coupon

 docker exec cli peer chaincode invoke -C <channelname> -n <chaincodename> -c '{"Args":["createCoupon", "Big Sale", "10-07-2019", "31-12-2019", "10.5", "Ankit1","Ankit1@gmail.com"]}'

Get Coupon by CouponID

docker exec cli peer chaincode invoke -C <channelname> -n <chaincodename> -c '{"Args":["getCouponById","<couponID>"]}'
