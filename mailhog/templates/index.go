package templates

func Index() string {
	return `
<style>
  .messages {
    height: 30%;
    overflow-y: scroll;
  }
  .preview {
    height: 70%;
    border-top: 1px solid #CCCCCC;
  }
  .preview #headers {
    border-bottom: 1px solid #DDDDDD;
  }
  .selected {
    background: #DADAFA;
  }
  table tbody {
    overflow: scroll;
  }
  table td {
    padding: 2px 4px 2px 4px !important;
  }
</style>
<div class="modal fade" id="confirm-delete-all">
  <div class="modal-dialog">
    <div class="modal-content">
      <div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>
        <h4 class="modal-title">Delete all messages?</h4>
      </div>
      <div class="modal-body">
        <p>Are you sure you want to delete all messages?</p>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
        <button type="button" class="btn btn-danger" ng-click="deleteAllConfirm()">Delete all messages</button>
      </div>
    </div>
  </div>
</div>

<div class="messages">
  <table class="table">
    <tr>
      <th>From</th>
      <th>To</th>
      <th>Subject</th>
      <th>Received</th>
      <th>Actions</th>
    </tr>
    <tbody>
      <tr ng-repeat="message in messages" ng-click="selectMessage(message)" ng-class="{ selected: message == preview }">
        <td>
          {{ message.From.Mailbox }}@{{ message.From.Domain }}
        </td>
        <td>
          <span ng-repeat="to in message.To">
            {{ to.Mailbox }}@{{ to.Domain }}
          </span>
        </td>
        <td>
          {{ message.Content.Headers.Subject }}
        </td>
        <td>
          {{ date(message.Created) }}
        </td>
        <td>
          <button class="btn btn-xs btn-default" title="Delete" ng-click="deleteOne(message)"><span class="glyphicon glyphicon-remove"></span></button>
        </td>
      </tr>
    </tbody>
  </table>
</div>
<div class="preview">
  <table class="table" id="headers">
    <tr ng-repeat="(header, value) in preview.Content.Headers">
      <td>
        {{ header }}
      </td>
      <td>
        {{ value }}
      </td>
    </tr>
  </table>
  {{ preview.Content.Body }}
</div>
`;
}