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

void create_connection()
{
    int listenSock, connSock, ret, in, flags, i;
    struct sctp_initmsg initmsg;
    struct sctp_event_subscribe event = {0};

    ue_info_arr = (struct ue_info *)malloc(num_ue * sizeof(struct ue_info));
    socket_fd = socket(AF_INET, SOCK_SEQPACKET, IPPROTO_SCTP);
    if (socket_fd == -1)
    {
        printf("Failed to create socket\n");
        perror("socket()");
        exit(1);
    }
    char *szAddress;
    szAddress = "127.0.0.1";
    int iPort = 5001;
    ret = setsockopt(socket_fd, IPPROTO_SCTP, SCTP_EVENTS, &event, sizeof(struct sctp_event_subscribe));
    bzero((void *)&servaddr, sizeof(servaddr));
    servaddr.sin_family = AF_INET;
    servaddr.sin_addr.s_addr = inet_addr(szAddress);
    servaddr.sin_port = htons(iPort);

    ret = bind(socket_fd, (struct sockaddr *)&servaddr, sizeof(struct sockaddr_in));
    if (ret == -1)
    {
        printf("Bind failed \n");
        perror("bind()");
        close(socket_fd);
        exit(1);
    }
    ret = listen(socket_fd, 1);
    if (ret == -1)
    {
        printf("listen() failed \n");
        perror("listen()");
        close(socket_fd);
        exit(1);
    }

}
void create_connection1()
{
        int SctpScocket, n, flags;
        socklen_t from_len;

        struct sockaddr_in addr = {0};
        struct sctp_sndrcvinfo sinfo = {0};
        struct sctp_event_subscribe event = {0};
        //char pRecvBuffer[RECVBUFSIZE + 1] = {0};

        char *szAddress;
        int iPort;
        char *szMsg;
        int iMsgSize;
        ue_info_arr = (struct ue_info *)malloc(num_ue * sizeof(struct ue_info));
        //get the arguments
        szAddress = "127.0.0.1";
        iPort = MY_PORT_NUM;
        szMsg = "Hello Client";
        iMsgSize = strlen(szMsg);
        if (iMsgSize > 1024)
        {
                printf("Message is too big for this test\n");
                return;
        }

        //here we may fail if sctp is not supported
        socket_fd = socket(AF_INET, SOCK_SEQPACKET, IPPROTO_SCTP);
        printf("socket created...\n");

        //make sure we receive MSG_NOTIFICATION
        setsockopt(socket_fd, IPPROTO_SCTP, SCTP_EVENTS, &event, sizeof(struct sctp_event_subscribe));
        printf("setsockopt succeeded...\n");

        addr.sin_family = AF_INET;
        addr.sin_port = htons(iPort);
        addr.sin_addr.s_addr = inet_addr(szAddress);

        //bind to specific server address and port
        bind(socket_fd, (struct sockaddr *)&addr, sizeof(struct sockaddr_in));
        printf("bind succeeded...\n");

        //wait for connections
        listen(socket_fd, 1);
        printf("listen succeeded...\n");

}
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

    //create_connection();
    create_connection1();
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
                        //printf("errno: %d\n", errno);
                        perror("Description: ");
                        sleep(1);
                        continue;
                }
        else
        {
            //Add '\0' in case of text data
            //buffer[in] = '\0';

            printf(" Length of Data received: %d\n", in);
            printf(" Data : %s\n", (char *)buffer);
           /* char *buffer = "hello";
            n = sctp_sendmsg(socket_fd, (const void *)buffer, 5, (struct sockaddr *)&addr, from_len, htonl(MY_PORT_NUM), 0, 0, 0, 0);
if (n == -1)
    {
        printf("Error in sctp_sendmsg\n");
        perror("sctp_sendmsg()");
        //return;
    } else {
        printf("Sent\n");
    }*/
            handle_received_message(buffer);
        }
        //close (socket_fd);
    }
    /*ret = wait_for_msg();
    if (ret == -1)
        goto clean;*/
    getchar();
clean:
    cleanup();
    return 0;
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
        } /*else {
            //    id = msg->attach_req.mme_ue_s1ap_id;
            //else
        }*/
        //ue_info_arr[id].socket = connSock;
        process_message(msg, id);
    }
    else
    { // Send buffer to mme function over http
    }
}
int wait_for_msg()
{
    int in, flags;
    struct sctp_sndrcvinfo sndrcvinfo;

    uint8_t buffer[MAX_BUFFER + 1];
    while (1)
    {
        in = sctp_recvmsg(socket_fd, buffer, sizeof(buffer), (struct sockaddr *)NULL,
                          0, &sndrcvinfo, &flags);
        if (in <= 0)
        {
            printf("Error in sctp_recvmsg\n");
            perror("sctp_recvmsg()");
            //continue;
            close(socket_fd);
            return -1;
        }
        else
        {
            printf("Got message by sctp_recvmsg, in:%d\n", in);
            handle_received_message(buffer);
        }
    }
    return 0;
}
void send_sctp_message(long id)
{
    int ret;

    ret = sctp_sendmsg(socket_fd, (const void *)&ue_info_arr[id].message, (size_t)ue_info_arr[id].datalen, (struct sockaddr *)&addr, from_len, htonl(MY_PORT_NUM), 0, 0 /*stream 0*/, 0, 0);

    if (ret == -1)
    {
        printf("Error in sctp_sendmsg\n");
        perror("sctp_sendmsg()");
        return;
    }
    else
        printf("Successfully sent %d bytes data to server\n", ret);
}
void cleanup()
{
    close(socket_fd);
    free(ue_info_arr);
}
void process_message(struct message *msg, long id)
{
    switch (msg->type)
    {
    case ATTACH_REQ:
    {
        struct attach_req *attach_req = (struct attach_req *)&msg->message_union;
        //int id = ue_info_arr[id].mme_ue_s1ap_id = id;
        ue_info_arr[id].ue_id[0] = attach_req->imsi[0];
        ue_info_arr[id].tai = attach_req->tai;
        ue_info_arr[id].enb_ue_s1ap_id = attach_req->enb_ue_s1ap_id;
        ue_info_arr[id].ue_state = IDLE;
        build_auth_request(id);
        send_sctp_message(id);
    }
    break;

    case AUTH_RES:
    {
        struct sec_mode_command *sec_mode_command = (struct sec_mode_command *)&msg->message_union;
        int id = sec_mode_command->mme_ue_s1ap_id;
        build_sec_mode_command(id);
    }
    break;

    case SEC_MODE_COMPLETE:
        build_attach_accept(id);
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
    ue_info_arr[id].message.message_union.sec_mode_command.mme_ue_s1ap_id = id;
    ue_info_arr[id].message.message_union.sec_mode_command.enb_ue_s1ap_id = ue_info_arr[id].enb_ue_s1ap_id;
    ue_info_arr[id].message.message_union.sec_mode_command.sec_algo = 0xaa;
}

void build_attach_accept(long id)
{
}