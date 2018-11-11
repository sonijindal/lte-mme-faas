#include <netinet/in.h>
#include <netinet/sctp.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <time.h>
#include <unistd.h>

#include "common.h"
#include "ue.h"
#include "mme.h"

struct sockaddr_in addr = {0};
socklen_t from_len;

int main(int argc, char *argv[])
{
    int flags, in;
    uint8_t buffer[MAX_BUFFER + 1] = {0};
    struct sctp_sndrcvinfo sndrcvinfo = {0};

    if (argc != 2)
    {
        printf("Usage: ./mme <num_ue>\n");
        return 0;
    }
    num_ue = atoi(argv[1]);

    create_connection();
    while (1)
    {
        struct sctp_sndrcvinfo sinfo = {0};
        int n;
        bzero(buffer, MAX_BUFFER + 1);
        flags = 0;
        memset((void *)&addr, 0, sizeof(struct sockaddr_in));
        from_len = (socklen_t)sizeof(struct sockaddr_in);
        memset((void *)&sinfo, 0, sizeof(struct sctp_sndrcvinfo));

        n = sctp_recvmsg(socket_fd, (void *)buffer, MAX_BUFFER, (struct sockaddr *)&addr, &from_len, &sinfo, &flags);
        if (-1 == n)
        {
            printf("Error with sctp_recvmsg: -1... waiting\n");
            perror("Description: ");
            sleep(1);
            continue;
        }
        else
        {
            handle_received_message(buffer);
        }
    }

    getchar();
clean:
    cleanup();
    return 0;
}
void create_connection()
{
    int SctpScocket, n, flags;
    socklen_t from_len;

    struct sockaddr_in addr = {0};
    struct sctp_sndrcvinfo sinfo = {0};
    struct sctp_event_subscribe event = {0};

    char *szAddress;
    int iPort;

    int iMsgSize;
    ue_info_arr = (struct ue_info *)malloc(num_ue * sizeof(struct ue_info));
    //get the arguments
    szAddress = "127.0.0.1";
    iPort = MY_PORT_NUM;

    //here we may fail if sctp is not supported
    socket_fd = socket(AF_INET, SOCK_SEQPACKET, IPPROTO_SCTP);
    debug_print("socket created...\n");

    //make sure we receive MSG_NOTIFICATION
    setsockopt(socket_fd, IPPROTO_SCTP, SCTP_EVENTS, &event, sizeof(struct sctp_event_subscribe));
    debug_print("setsockopt succeeded...\n");

    addr.sin_family = AF_INET;
    addr.sin_port = htons(iPort);
    addr.sin_addr.s_addr = inet_addr(szAddress);

    //bind to specific server address and port
    bind(socket_fd, (struct sockaddr *)&addr, sizeof(struct sockaddr_in));
    debug_print("bind succeeded...\n");

    //wait for connections
    listen(socket_fd, 1);
    debug_print("listen succeeded...\n");
}

void handle_received_message(uint8_t *buffer)
{
    if (handle_local)
    {
        long id = 0;
        struct message *msg = (struct message *)buffer;
        if (msg->type == ATTACH_REQ)
        {
            ue_info_arr[id].mme_ue_s1ap_id = id;
            ++id;
        }
        process_message(msg, id);
    }
    else
    { // TODO: Send buffer to mme function over http
    }
}

void process_message(struct message *msg, long id)
{
    switch (msg->type)
    {
    case ATTACH_REQ:
    {
        debug_print("Attach Req received\n");
        struct attach_req *attach_req = (struct attach_req *)&msg->message_union;
        ue_info_arr[id].ue_id[0] = attach_req->imsi[0];
        ue_info_arr[id].tai = attach_req->tai;
        ue_info_arr[id].enb_ue_s1ap_id = attach_req->enb_ue_s1ap_id;
        ue_info_arr[id].ue_state = IDLE;
        ue_info_arr[id].plmn_id = attach_req->plmn_id;
        build_auth_request(id);
        send_sctp_message(id);
    }
    break;

    case AUTH_RES:
    {
        debug_print("Authentication response received\n");
        struct sec_mode_command *sec_mode_command = (struct sec_mode_command *)&msg->message_union;
        int id = sec_mode_command->mme_ue_s1ap_id;
        build_sec_mode_command(id);
        send_sctp_message(id);
    }
    break;

    case SEC_MODE_COMPLETE:
    {
        debug_print("Security Mode Complete received\n");
        struct sec_mode_complete *sec_mode_complete = (struct sec_mode_complete *)&msg->message_union;
        int id = sec_mode_complete->mme_ue_s1ap_id;
        ue_info_arr[id].ue_state = CONNECTED;
        build_attach_accept(id);
        send_sctp_message(id);
    }
    break;

    default:
        printf("Unknown message received!\n");
    break;
    }
}

void build_auth_request(long id)
{
    ue_info_arr[id].message.type = AUTH_REQ;
    ue_info_arr[id].message.message_union.auth_req.mme_ue_s1ap_id = id;
    ue_info_arr[id].message.message_union.auth_req.enb_ue_s1ap_id = ue_info_arr[id].enb_ue_s1ap_id;
    ue_info_arr[id].message.message_union.auth_req.auth_challenge = 0xaa;
    ue_info_arr[id].datalen = sizeof(struct message);
}

void build_sec_mode_command(long id)
{
    ue_info_arr[id].message.type = SEC_MODE_COMMAND;
    ue_info_arr[id].message.message_union.sec_mode_command.mme_ue_s1ap_id = id;
    ue_info_arr[id].message.message_union.sec_mode_command.enb_ue_s1ap_id = ue_info_arr[id].enb_ue_s1ap_id;
    ue_info_arr[id].message.message_union.sec_mode_command.sec_algo = 0xaa;
    ue_info_arr[id].datalen = sizeof(struct message);
}

void build_attach_accept(long id)
{
    ue_info_arr[id].message.type = ATTACH_ACCEPT;
    ue_info_arr[id].message.message_union.attach_accept.mme_ue_s1ap_id = ue_info_arr[id].mme_ue_s1ap_id;
    ue_info_arr[id].message.message_union.attach_accept.ambr = 100;
    ue_info_arr[id].message.message_union.attach_accept.sec_cap = 100;
    ue_info_arr[id].datalen = sizeof(struct message);
}

void send_sctp_message(long id)
{
    int ret;

    ret = sctp_sendmsg(socket_fd, (const void *)&ue_info_arr[id].message, (size_t)ue_info_arr[id].datalen,
                       (struct sockaddr *)&addr, from_len, htonl(MY_PORT_NUM), 0, 0, 0, 0);

    if (ret == -1)
    {
        printf("Error in sctp_sendmsg\n");
        perror("sctp_sendmsg()");
        return;
    }
    else
        debug_print("Successfully sent %d bytes data to ENB\n", ret);
}
void cleanup()
{
    close(socket_fd);
    free(ue_info_arr);
}