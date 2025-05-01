package extproc

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	healthPb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/structpb"

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

		phase := req.GetMetadataContext().GetFilterMetadata()["phase"]

		stage := ""
		if phase == nil {
			log.Printf("no phase found in metadata")
			stage = "start"
		} else {
			stage = phase.GetFields()["stage"].GetStringValue()
			stage = stage + " - next"
		}

		log.Printf("Stage: %s\n", stage)

		resp := &extProcPb.ProcessingResponse{
			DynamicMetadata: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"phase": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"stage": structpb.NewStringValue(stage),
						},
					}),
				},
			},
			Response: &extProcPb.ProcessingResponse_RequestHeaders{
				RequestHeaders: &extProcPb.HeadersResponse{
					Response: &extProcPb.CommonResponse{
						HeaderMutation: &extProcPb.HeaderMutation{
							SetHeaders: []*configPb.HeaderValueOption{
								{
									Header: &configPb.HeaderValue{
										Key:   "x-went-into-req-headers",
										Value: "true",
									},
								},
							},
						},
					},
				},
			},
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
