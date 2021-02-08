package mock

import (
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
}

func NewGrpServer(host string, port string) *Server {
	server := &Server{}
	server.createServer(host, port)

	return server
}

func (server *Server) StartServer() {
	server.serve()
}

func (server *Server) createServer(host string, port string) {
	listener, error := net.Listen("tcp", host+":"+port)
	if error != nil {
		FwLog.Error(error, "failed to listen")
	} else {
		server.listener = listener
		server.grpcServer = grpc.NewServer()
		FwLog.Info("grpc server has been created successfully")
	}
}

func (server *Server) serve() {
	FwLog.Info("grpc Server is listening...")

	if error := server.grpcServer.Serve(server.listener); error != nil {
		FwLog.Error(error, "failed start grpc server")
	}
	FwLog.Info("grpc Server stopped...")

}

func (server *Server) GetServer() *grpc.Server {
	return server.grpcServer
}
