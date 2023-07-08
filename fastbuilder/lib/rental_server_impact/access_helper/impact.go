package access_helper

import (
	"context"
	"fmt"
	"phoenixbuilder/fastbuilder/core"
	fbauth "phoenixbuilder/fastbuilder/cv4/auth"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/bundle"
	neomega_core "phoenixbuilder/fastbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/fastbuilder/lib/rental_server_impact/challenges"
	"phoenixbuilder/fastbuilder/lib/rental_server_impact/info_collect_utils"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
)

func ImpactServer(ctx context.Context, options *Options) (conn *minecraft.Conn, omegaCore *bundle.MicroOmega, deadReason chan error, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if options.MaximumWaitTime > 0 {
		ctx, _ = context.WithTimeout(ctx, options.MaximumWaitTime)
	}
	clientOptions := fbauth.MakeDefaultClientOptions()
	clientOptions.AuthServer = options.AuthServer
	fmt.Println("connecting to fb server...")
	fbClient := fbauth.CreateClient(clientOptions)
	fmt.Println("done connecting to fb server")
	if options.FBUserToken == "" {
		var err_val string
		fmt.Println("obtaining fb token from fb server...")
		options.FBUserToken, err_val = fbClient.GetToken(options.FBUsername, options.FBUserPassword)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%v: %s", ErrFBUserCenterLoginFail, err_val)
		}
		fmt.Println("done obtaining fb token from fb server")
	}
	if options.WriteBackToken {
		info_collect_utils.WriteFBToken(options.FBUserToken, info_collect_utils.LoadTokenPath())
	}
	authenticator := fbauth.NewAccessWrapper(fbClient, options.ServerCode, options.ServerPassword, options.FBUserToken)
	{
		connectMCServer := func() (conn *minecraft.Conn, err error) {
			connectCtx := ctx
			if options.ServerConnectionTimeout != 0 {
				connectCtx, _ = context.WithTimeout(ctx, options.ServerConnectionTimeout)
			}
			conn, err = core.InitializeMinecraftConnection(connectCtx, authenticator)
			if err != nil {
				if connectCtx.Err() != nil {
					return nil, ErrRentalServerConnectionTimeOut
				}
				return nil, fmt.Errorf("%v :%v", ErrFailToConnectRentalServer, err)
			}
			return conn, nil
		}
		fmt.Println("connecting to mc server...")
		retryTimes := 0
		for {
			conn, err = connectMCServer()
			if err == nil {
				break
			} else {
				fmt.Println(err)
			}
			if options.ServerConnectRetryTimes <= 0 {
				break
			}
			retryTimes++
			fmt.Printf("fail connecting to mc server, retrying: %v\n", retryTimes)
			options.ServerConnectRetryTimes--
		}
		if err != nil {
			return nil, nil, nil, err
		}
		fmt.Println("done connecting to mc server")
	}
	omegaCore = bundle.NewMicroOmega(neomega_core.NewInteractCore(conn), func() omega.MicroUQHolder {
		return uqholder.NewMicroUQHolder(conn)
	}, options.MicroOmegaOption)
	deadReason = make(chan error)
	challengeSolver := challenges.NewPyRPCResponder(omegaCore, fbClient.Uid,
		fbClient.TransferData,
		fbClient.TransferCheckNum,
	)
	go func() {
		options.ReadLoopFunction(conn, deadReason, omegaCore)
	}()
	{
		fmt.Println("coping with rental server challenges ...")
		challengeSolvingCtx := ctx
		if options.ChallengeSolvingTimeout != 0 {
			challengeSolvingCtx, _ = context.WithTimeout(ctx, options.ChallengeSolvingTimeout)
		}
		success := challengeSolver.ChallengeCompete(challengeSolvingCtx)
		if !success {
			return nil, nil, nil, ErrFBChallengeSolvingTimeout
		}
		fmt.Println("done coping with rental server challenges")
	}
	if options.ReasonWithPrivilegeStuff {
		fmt.Printf("checking bot op permission and game cheat mode...\n")
		helper := challenges.NewOperatorChallenge(omegaCore, func() {
			if options.OpPrivilegeRemovedCallBack != nil {
				options.OpPrivilegeRemovedCallBack()
			}
			if options.DieOnLosingOpPrivilege {
				deadReason <- ErrBotOpPrivilegeRemoved
			}
		})
		waitErr := make(chan error)
		go func() {
			waitErr <- helper.WaitForPrivilege(ctx)
		}()
		select {
		case err = <-waitErr:
		case err = <-deadReason:
		}
		if err != nil {
			return nil, nil, nil, err
		}
		fmt.Printf("done checking bot op permission and game cheat mode\n")
	}
	if options.MakeBotCreative {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamemode c @s", func(output *packet.CommandOutput) {
			fmt.Printf("done setting bot to creative mode\n")
		})
	}
	if options.DisableCommandBlock {
		omegaCore.GetGameControl().SendPlayerCmdAndInvokeOnResponseWithFeedback("gamerule commandblocksenabled false", func(output *packet.CommandOutput) {
			fmt.Printf("done setting commandblocksenabled false\n")
		})
	}
	return conn, omegaCore, deadReason, nil
}
