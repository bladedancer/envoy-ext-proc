package extproc

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	typePb "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}
type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *healthPb.HealthCheckRequest) (*healthPb.HealthCheckResponse, error) {
	log.Printf("Handling grpc Check request + %s", in.String())
	return &healthPb.HealthCheckResponse{Status: healthPb.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *healthPb.HealthCheckRequest, srv healthPb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

// Demo Ext-Proc server
func (s *server) Process(srv extProcPb.ExternalProcessor_ProcessServer) error {
	log.Println(" ")
	log.Println(" ")
	log.Println("Started process:  -->  ")

	ctx := srv.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		log.Println(" ")
		log.Println(" ")
		log.Println("Got stream:  -->  ")

		resp := &extProcPb.ProcessingResponse{}

		switch v := req.Request.(type) {

		case *extProcPb.ProcessingRequest_RequestHeaders:

			log.Println("--- In RequestHeaders processing ...")
			// r := req.Request
			// h := r.(*extProcPb.ProcessingRequest_RequestHeaders)

			// log.Printf("Request: %+v\n", r)
			// log.Printf("Headers: %+v\n", h)
			// log.Printf("EndOfStream: %v\n", h.RequestHeaders.EndOfStream)

			if rand.Uint32N(100) > config.SuccessPercentage {
				log.Println("Simulating failure")
				return status.Error(codes.Internal, "Simulated failure")
			}

			resp = &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_ImmediateResponse{
					ImmediateResponse: &extProcPb.ImmediateResponse{
						Status: &typePb.HttpStatus{
							Code: 200,
						},
						Body: []byte("Hello from the external processor"),
						GrpcStatus: &extProcPb.GrpcStatus{
							Status: 0,
						},
					},
				},
			}

		default:
			log.Printf("Unhandled Request type %+v\n", v)
		}

		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}

// Run entry point for Envoy XDS command line.
func Run() error {
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
	if err != nil {
		log.Fatal(err)
	}

	extProcPb.RegisterExternalProcessorServer(grpcServer, &server{})
	healthPb.RegisterHealthServer(grpcServer, &healthServer{})

	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	log.Infof("Listening on %d", config.Port)

	// Wait for CTRL-c shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done

	grpcServer.GracefulStop()
	log.Info("Shutdown")
	return nil
}
