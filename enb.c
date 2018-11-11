#include <netinet/in.h>
#include <netinet/sctp.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

#include "common.h"
#include "enb.h"
#include "ue.h"

int main(int argc, char *argv[])
{
	pthread_t thread_id;
	if (argc != 2)
	{
		printf("Usage: ./enb <num_ue>\n");
		return 0;
	}
	num_ue = atoi(argv[1]);
	bzero((void *)&servaddr, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_port = htons(MY_PORT_NUM);
	servaddr.sin_addr.s_addr = inet_addr("127.0.0.1");

	ue_info_arr = (struct ue_info *)malloc(num_ue * sizeof(struct ue_info));

	create_connection();
	sleep(1);
	pthread_create(&thread_id, NULL, wait_for_msg, NULL);
	for (long id = 0; id < num_ue; ++id)
	{
		process_message(NULL, id);
	}
	getchar();
	cleanup();
	pthread_join(thread_id, NULL);

	return 0;
}

void *wait_for_msg(void *vargp)
{
	int in, flags;
	struct sctp_sndrcvinfo sndrcvinfo;
	while (1)
	{
		uint8_t buffer[MAX_BUFFER + 1];
		in = sctp_recvmsg(socket_fd, buffer, sizeof(buffer),
						  (struct sockaddr *)NULL, 0, &sndrcvinfo, &flags);
		if (in <= 0)
		{
			printf("Error in sctp_recvmsg\n");
			perror("sctp_recvmsg()");
			close(socket_fd);
			return NULL;
		}
		else
		{
			handle_received_message(buffer);
		}
	}
	return NULL;
}
void handle_received_message(uint8_t *buffer)
{
	struct message *msg = (struct message *)buffer;
	process_message(msg, 0);
}
void send_sctp_message(long id)
{
	int ret;
	ret = sctp_sendmsg(socket_fd, (void *)&ue_info_arr[id].message,
					   (size_t)ue_info_arr[id].datalen, NULL, 0, 0, 0, 0, 0, 0);
	if (ret == -1)
	{
		printf("Error in sctp_sendmsg\n");
		perror("sctp_sendmsg()");
	}
	else
		debug_print("Successfully sent %d bytes data to MME\n", ret);
}
void cleanup()
{
	close(socket_fd);
	free(ue_info_arr);
}
void create_connection()
{
	int ret = -1;
	struct sctp_initmsg initmsg = {0};
	socket_fd = socket(AF_INET, SOCK_STREAM, IPPROTO_SCTP);
	// ue_info_arr[id].socket = socket_fd;
	if (socket_fd == -1)
	{
		printf("Socket creation failed\n");
		perror("socket()");
		exit(1);
	}
	// set the association options
	initmsg.sinit_num_ostreams = 1;
	setsockopt(socket_fd, IPPROTO_SCTP, SCTP_INITMSG, &initmsg, sizeof(initmsg));
	debug_print("setsockopt succeeded...\n");

	ret = connect(socket_fd, (struct sockaddr *)&servaddr, sizeof(servaddr));

	if (ret == -1)
	{
		printf("Connection failed\n");
		perror("connect()");
		close(socket_fd);
		exit(1);
	}
}

void process_message(struct message *msg, long id)
{
	if (msg == NULL)
	{
		// First message for this UE, so send attach
		build_attach(id);
		send_sctp_message(id);
	}
	else
	{
		switch (msg->type)
		{
		case AUTH_REQ:
		{
			debug_print("Auth Req received\n");
			struct auth_req *auth_req = (struct auth_req *)&msg->message_union;
			int id = auth_req->enb_ue_s1ap_id;
			ue_info_arr[id].message.message_union.auth_res.mme_ue_s1ap_id =
				auth_req->mme_ue_s1ap_id;
			ue_info_arr[id].message.message_union.auth_res.auth_challenge_answer =
				auth_req->auth_challenge;
			build_auth_response(id);
			send_sctp_message(id);
		}
		break;

		case SEC_MODE_COMMAND:
		{
			debug_print("Security Mode Command received\n");
			struct sec_mode_command *sec_mode_command =
				(struct sec_mode_command *)&msg->message_union;
			int id = sec_mode_command->enb_ue_s1ap_id;
			ue_info_arr[id].message.message_union.sec_mode_complete.mme_ue_s1ap_id = sec_mode_command->mme_ue_s1ap_id;
			build_sec_mode_complete(id);
			send_sctp_message(id);
		}
		break;

		case ATTACH_ACCEPT:
			printf("Attach Accept received\n");
		break;
		
		default:
			printf("Unknown message received!\n");
		break;
		}
	}
}

void build_attach(long id)
{
	ue_info_arr[id].message.type = ATTACH_REQ;
	ue_info_arr[id].message.message_union.attach_req.enb_ue_s1ap_id =
		id; // enb_ue_s1ap_id uniquely identifies the UE at enb.
	ue_info_arr[id].message.message_union.attach_req.imsi[0] = 1;
	ue_info_arr[id].message.message_union.attach_req.tai = 1;
	ue_info_arr[id].message.message_union.attach_req.net_cap = 1;
	ue_info_arr[id].message.message_union.attach_req.plmn_id = 1;
	ue_info_arr[id].datalen = sizeof(struct message);
}

void build_auth_response(long id)
{
	ue_info_arr[id].message.type = AUTH_RES;
	ue_info_arr[id].message.message_union.auth_res.enb_ue_s1ap_id = id;
	ue_info_arr[id].datalen = sizeof(struct message);
}

void build_sec_mode_complete(long id)
{
	ue_info_arr[id].message.type = SEC_MODE_COMPLETE;
	ue_info_arr[id].message.message_union.sec_mode_complete.enb_ue_s1ap_id = id;
	ue_info_arr[id].message.message_union.sec_mode_complete.tai = ue_info_arr[id].tai;
	ue_info_arr[id].message.message_union.sec_mode_complete.plmn_id = ue_info_arr[id].plmn_id;
	ue_info_arr[id].datalen = sizeof(struct message);
}