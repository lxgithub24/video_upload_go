package rpcClient

import (
	"context"
	"golivephoto/config"
	pb "golivephoto/grpcService/pb"
	"strconv"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

func getRegisterRPC(rpcname string) string {
	Client, err := api.NewClient(&api.Config{Address: config.ConsulConfig}) //local
	if err != nil {
		panic(err)
	}
	addr, querymeta, err := Client.Catalog().Service(rpcname, "", nil)
	if err != nil {
		println(querymeta)
		return ""
	}
	rpcUrl := addr[0].ServiceAddress + ":" + strconv.Itoa(addr[1].ServicePort)
	return rpcUrl
}

func Get_livephoto_gcid(stub string) []string {
	rpcUrl := getRegisterRPC(config.UploadStubRPC)
	conn, err := grpc.Dial(rpcUrl, grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	c := pb.NewUploadStubClient(conn)
	r, err := c.QueryStubInfo(context.Background(), &pb.GetRequest{Stub: stub})
	if err != nil {
		panic(err)
	}
	var result []string
	for _, v := range r.Results {
		result = append(result, v.FileGcid)
	}
	return result
}
